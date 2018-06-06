package main

import (
	"encoding/json"
	"fmt"
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

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func executeWorkload(w http.ResponseWriter, req *http.Request) {
	// workloadDetails := types.WorkloadDetails{}
	step := types.Step{}
	// helpers.PanicOnErrorAPI(helpers.ParseRequest(req, &workloadDetails), w)
	helpers.PanicOnErrorAPI(helpers.ParseRequest(req, &step), w)
	// Fetch yaml from predefined_steps table
	rdbSession, err := r.Connect(r.ConnectOpts{
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
	})
	helpers.PanicOnErrorAPI(err, w)
	// log.Println(workloadDetails)
	cursor, err := r.Table("predefined_steps").Filter(map[string]interface{}{
		// "name": workloadDetails.Task,
		"name": step.StepTemplate,
	}).Run(rdbSession)
	helpers.PanicOnErrorAPI(err, w)
	defer cursor.Close()
	var yamlFromDB types.YAMLFromDB
	err = cursor.One(&yamlFromDB)
	helpers.PanicOnErrorAPI(err, w)
	// Workload name of the format - {{.Repo}}-{{.PRID}}-{{.SrcTopCommit}}-{{.Task}}
	// workloadName := workloadDetails.Repo +
	// 	"-" + strconv.Itoa(workloadDetails.PRID) +
	// 	"-" + workloadDetails.SrcTopCommit +
	// 	"-" + workloadDetails.Task +
	// 	"-" + strconv.Itoa(workloadDetails.StepID)

	// workloadName := step.UniqueKey
	workloadPath := "./yaml/" + step.UniqueKey + ".yaml"
	fileContentInBytes := replaceVariables(yamlFromDB, step, workloadPath)
	helpers.PanicOnErrorAPI(err, w)
	// log.Printf("here - %s", string(fileContentInBytes))
	err = ioutil.WriteFile(workloadPath, fileContentInBytes, 0777)
	helpers.PanicOnErrorAPI(err, w)
	go runKubeCTL(step.UniqueKey, workloadPath, step.UniqueKey)
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
	// configBuf := new(bytes.Buffer)
	// log.Println(yfr.Config)
	// log.Printf("%+v", step)
	// tmpl := template.Must(template.New(workloadPath).Parse(yfr.Config))
	// err := tmpl.Execute(configBuf, step)
	// if err != nil {
	// 	return nil, err
	// }
	// return configBuf.Bytes(), nil
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

func runKubeCTL(uniqueKey, workloadPath, uniqueID string) {
	resChan := make(chan types.WorkloadResult)
	go func(uniqueID string) {
		for {
			select {
			case wr := <-resChan:
				_, err := helpers.Post("http://localhost:5500/callback", wr, nil)
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}(uniqueID)
	defer close(resChan)
	cmd := exec.Command("kubectl", "--kubeconfig", *ConfigPath, "create", "-f", workloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		resChan <- types.WorkloadResult{
			ID:      uniqueID,
			Result:  "Failed",
			Details: err.Error(),
		}
		return
	}
	log.Println("yaml executed")
	// Poll Kube API for result of workload
	log.Println("job-name=" + uniqueKey)
	listOpts := metav1.ListOptions{
		LabelSelector: "job-name=" + uniqueKey,
	}
	log.Println("listopts done")
	time.Sleep(time.Duration(2 * time.Second))
	watcherI, err := Clientset.BatchV1().Jobs("default").Watch(listOpts)
	if err != nil {
		resChan <- types.WorkloadResult{
			ID:      uniqueID,
			Result:  "Failed",
			Details: err.Error(),
		}
		return
	}
	log.Println("Created watcherI")
	ch := watcherI.ResultChan()
	defer watcherI.Stop()
	for {
		select {
		case event := <-ch:
			job, isPresent := event.Object.(*batchv1.Job)
			if !isPresent {
				log.Println("Unknown Object Type")
				resChan <- types.WorkloadResult{
					ID:      uniqueID,
					Result:  "Failed",
					Details: "Unknown Object Type",
				}
				return
			}
			log.Printf("Job: %s -> Active: %d, Succeeded: %d, Failed: %d",
				job.Name, job.Status.Active, job.Status.Succeeded, job.Status.Failed)
			switch event.Type {
			case watch.Modified:
				sendResponse := false
				log.Println("New modification poll")
				res := "Failed"
				errMsg := "Job Failed on K8s"
				if job.Status.Failed > 0 {
					log.Println("Workload Failed")
					sendResponse = true
				}
				if job.Status.Active == 0 {
					if job.Status.Succeeded == 1 {
						res = "Succeeded"
						errMsg = ""
					}
					log.Println("Workload Succeeded")
					sendResponse = true
				}
				if sendResponse {
					log.Println("Stopping Poll")
					resChan <- types.WorkloadResult{
						ID:      uniqueID,
						Result:  res,
						Details: errMsg,
					}
					return
				}
			}
		}
	}
}
