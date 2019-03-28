package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"

	r "gopkg.in/gorethink/gorethink.v4"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func executeWorkload(w http.ResponseWriter, req *http.Request) {
	step := types.Step{}
	helpers.PanicOnErrorAPI(helpers.ParseRequest(req, &step), w)
	// Fetch yaml from predefined_steps table
	rdbSession, err := r.Connect(r.ConnectOpts{
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
	})
	helpers.PanicOnErrorAPI(err, w)
	defer rdbSession.Close()
	cursor, err := r.Table("predefined_steps").Filter(map[string]interface{}{
		"name": step.StepTemplate,
	}).Run(rdbSession)
	helpers.PanicOnErrorAPI(err, w)
	defer cursor.Close()
	var yamlFromDB types.YAMLFromDB
	err = cursor.One(&yamlFromDB)
	helpers.PanicOnErrorAPI(err, w)

	workloadPath := "./yaml/" + step.UniqueKey + ".yaml"
	fileContentInBytes := replaceVariables(yamlFromDB, step, workloadPath)
	helpers.PanicOnErrorAPI(err, w)
	log.Println("**********************************")
	err = ioutil.WriteFile(workloadPath, fileContentInBytes, 0777)
	log.Println("----------------------------------")
	helpers.PanicOnErrorAPI(err, w)
	go runKubeCTL(step.UniqueKey, workloadPath)
	eRes := helpers.Response{
		State: "Workload triggered",
		Code:  200,
	}
	inBytes, _ := json.Marshal(eRes)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}

func replaceVariables(yamlFromDB types.YAMLFromDB, step types.Step, workloadPath string) []byte {
	// // Some replaces happen here
	fmt.Println(step.Replacers)
	for singleReplacer, value := range step.Replacers {
		yamlFromDB.Config = strings.Replace(yamlFromDB.Config, "{{."+singleReplacer+"}}", value, -1)
	}
	// Twice for overlapping substitution
	for singleReplacer, value := range step.Replacers {
		yamlFromDB.Config = strings.Replace(yamlFromDB.Config, "{{."+singleReplacer+"}}", value, -1)
	}
	fmt.Println(yamlFromDB.Config)
	return []byte(yamlFromDB.Config)
}

func runKubeCTL(uniqueKey, workloadPath string) {
	resChan := make(chan types.WorkloadResult)
	go func(uniqueKey string) {
		for {
			log.Println("Waiting on result channel for " + uniqueKey)
			select {
			case wr := <-resChan:
				log.Println("Sending floworch result for " + uniqueKey)
				_, err := helpers.Post("http://localhost:5500/callback", wr, nil)
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}(uniqueKey)
	defer close(resChan)
	cmd := exec.Command("kubectl", "--kubeconfig", *ConfigPath, "create", "-f", workloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		resChan <- types.WorkloadResult{
			UniqueKey: uniqueKey,
			Result:    types.FAILED,
			Details:   err.Error(),
		}
		return
	}
	log.Println("yaml executed")
	// Poll Kube API for result of workload
	log.Println("job-name=" + uniqueKey)
	// listOpts := metav1.ListOptions{
	// 	LabelSelector: "job-name=" + uniqueKey,
	// }
	log.Println("listopts done")
	time.Sleep(time.Duration(2 * time.Second))

	k8File, err := os.Open(workloadPath)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(k8File)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 2048)
	workloadTrackMap := make(map[string]int)

	watchChannel := make(chan types.WorkloadResult)
	defer close(watchChannel)
	for {
		ext := runtime.RawExtension{}
		if err := decoder.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		log.Println("-----------------------------")
		log.Println("raw: ", string(ext.Raw))

		versions := &runtime.VersionedObjects{}
		obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, versions)
		if err != nil {
			log.Fatal(err)
		}

		// GroupKind type
		workloadKind := gvk.GroupKind().Kind
		// non existent value is default which is 0
		workloadTrackMap[workloadKind]++

		objectKindI := obj.GetObjectKind()
		structuredObj := objectKindI.(*unstructured.Unstructured)
		labelSet := structuredObj.GetLabels()

		listOpts := metav1.ListOptions{
			LabelSelector: labels.Set(labelSet).String(),
		}

		log.Println(workloadKind)
		switch workloadKind {
		case "Job":
			go JobWatch(Clientset, watchChannel, "default", listOpts)
		case "StatefulSet":
			go StatefulSetWatch(Clientset, watchChannel, "default", listOpts)
		case "Service":
			go ServiceWatch(Clientset, watchChannel, "default", listOpts)
		default:
			log.Println("Unknown workload. Completed")
		}
	}

	totalWorkload := len(workloadTrackMap)
	receivedResults := 0

	/**
	 * result check receieved from go routines
	 */
	for {
		select {
		case event := <-watchChannel:
			log.Println("++++++++++++++++++++++++++Recieved Watch Event++++++++++++++++++++++++++++++")
			log.Println(event)
			receivedResults += 1
			if event.Result == types.FAILED {
				fmt.Println(event.Details)
				return
			}
			if workloadTrackMap[event.Kind] == 1 {
				delete(workloadTrackMap, event.Kind)
			} else {
				workloadTrackMap[event.Kind] -= 1
			}

			if receivedResults == totalWorkload && len(workloadTrackMap) == 0 {
				log.Println("Succesfully completed workload", uniqueKey)
				// hack for unique key specific stuff of floworch
				event.UniqueKey = uniqueKey
				resChan <- event
				return
			}

		}
	}

}
