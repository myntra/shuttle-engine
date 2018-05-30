package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"

	r "gopkg.in/gorethink/gorethink.v4"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func executeWorkload(w http.ResponseWriter, req *http.Request) {
	workloadDetails := types.WorkloadDetails{}
	helpers.PanicOnErrorAPI(helpers.ParseRequest(req, &workloadDetails), w)
	// Fetch yaml from predefined_steps table
	rdbSession, err := r.Connect(r.ConnectOpts{
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
	})
	helpers.PanicOnErrorAPI(err, w)
	// log.Println(workloadDetails)
	cursor, err := r.Table("predefined_steps").Filter(map[string]interface{}{
		"name": workloadDetails.Task,
	}).Run(rdbSession)
	helpers.PanicOnErrorAPI(err, w)
	defer cursor.Close()
	var yamlFromRethink types.YAMLFromRethink
	err = cursor.One(&yamlFromRethink)
	helpers.PanicOnErrorAPI(err, w)
	// Workload name of the format - {{.Repo}}-{{.PRID}}-{{.SrcTopCommit}}-{{.Task}}
	workloadName := workloadDetails.Repo +
		"-" + strconv.Itoa(workloadDetails.PRID) +
		"-" + workloadDetails.SrcTopCommit +
		"-" + workloadDetails.Task +
		"-" + strconv.Itoa(workloadDetails.StepID)
	workloadPath := "./yaml/" + workloadName + ".yaml"
	fileContentInBytes, err := replaceVariables(yamlFromRethink, workloadDetails, workloadPath)
	helpers.PanicOnErrorAPI(err, w)
	log.Printf("here - %s", string(fileContentInBytes))
	err = ioutil.WriteFile(workloadPath, fileContentInBytes, 0777)
	helpers.PanicOnErrorAPI(err, w)
	go runKubeCTL(workloadName, workloadPath, workloadDetails.WorkloadID)
	eRes := helpers.Response{
		State: "Workload triggered",
		Code:  200,
	}
	inBytes, _ := json.Marshal(eRes)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}

func replaceVariables(yfr types.YAMLFromRethink, wd types.WorkloadDetails, workloadPath string) ([]byte, error) {
	// Some replaces happen here
	configBuf := new(bytes.Buffer)
	log.Println(yfr.Config)
	log.Printf("%+v", wd)
	tmpl := template.Must(template.New(workloadPath).Parse(yfr.Config))
	err := tmpl.Execute(configBuf, wd)
	if err != nil {
		return nil, err
	}
	return configBuf.Bytes(), nil
}

func runKubeCTL(workloadName, workloadPath, workloadID string) {
	resChan := make(chan string)
	go func(workloadID string) {
		for {
			select {
			case res := <-resChan:
				callback(types.WorkloadResult{
					ID:     workloadID,
					Result: res,
				})
				return
			}
		}
	}(workloadID)
	defer close(resChan)
	cmd := exec.Command("kubectl", "--kubeconfig", *ConfigPath, "create", "-f", workloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	helpers.FailOnErr(err, resChan)
	log.Println("yaml executed")
	// Poll Kube API for result of workload
	log.Println("job-name=" + workloadName)
	listOpts := metav1.ListOptions{
		LabelSelector: "job-name=" + workloadName,
	}
	log.Println("listopts done")
	watcherI, err := Clientset.BatchV1().Jobs("default").Watch(listOpts)
	helpers.FailOnErr(err, resChan)
	log.Println("Created watcherI")
	ch := watcherI.ResultChan()
	defer watcherI.Stop()
	for {
		select {
		case event := <-ch:
			job, isPresent := event.Object.(*batchv1.Job)
			if !isPresent {
				log.Println("Unknown Object Type")
				continue
			}
			log.Printf("Job: %s -> Active: %d, Succeeded: %d, Failed: %d",
				job.Name, job.Status.Active, job.Status.Succeeded, job.Status.Failed)
			switch event.Type {
			case watch.Modified:
				log.Println("New modification poll")
				if job.Status.Active == 0 {
					res := "Failed"
					if job.Status.Succeeded == 1 {
						log.Println("Setting Succeeded")
						res = "Succeeded"
					}
					log.Println("Hitting API")
					resChan <- res
					log.Println("Stopping Poll")
					return
				}
			}
		}
	}
}

func callback(workloadResult types.WorkloadResult) {
	_, err := helpers.Post("http://localhost:5500/callback", workloadResult, nil)
	if err != nil {
		log.Println(err)
	}
}
