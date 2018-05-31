package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/ghodss/yaml"
	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
	r "gopkg.in/gorethink/gorethink.v4"
)

func executeHandler(w http.ResponseWriter, r *http.Request) {
	var workloadDetails types.WorkloadDetails
	err := helpers.ParseRequest(r, &workloadDetails)
	helpers.PanicOnErrorAPI(err, w)
	go InsertSteps(workloadDetails)
	eRes := types.ExecuteResponse{
		State:           "Workload triggered",
		WorkloadDetails: workloadDetails,
		Code:            200,
	}
	inBytes, err := json.Marshal(eRes)
	helpers.PanicOnErrorAPI(err, w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}

func procureContent(workloadDetails types.WorkloadDetails) []types.Step {
	rdbSession, err := r.Connect(r.ConnectOpts{
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
	})
	helpers.FailOnErr(err, nil)
	cursor, err := r.Table(workloadDetails.Stage + "_configs").Filter(map[string]interface{}{
		"id": workloadDetails.Repo + "-" + workloadDetails.DstBranch,
	}).Run(rdbSession)
	helpers.FailOnErr(err, nil)
	defer cursor.Close()
	var yamlFromRethink types.YAMLFromRethink
	err = cursor.One(&yamlFromRethink)
	helpers.FailOnErr(err, nil)
	// Replacing variables
	err = replaceVariables(&yamlFromRethink, workloadDetails)
	helpers.FailOnErr(err, nil)
	// Extracting yaml into json
	reg := regexp.MustCompile("- id:")
	matches := reg.FindAllStringIndex(yamlFromRethink.Config, -1)
	log.Println(yamlFromRethink.Config)
	stageSteps := make([]types.Step, len(matches))
	err = yaml.Unmarshal([]byte(yamlFromRethink.Config), &stageSteps)
	helpers.FailOnErr(err, nil)
	return stageSteps
}

func replaceVariables(yfr *types.YAMLFromRethink, wd types.WorkloadDetails) error {
	configBuf := new(bytes.Buffer)
	tmpl := template.Must(template.New("replaceFromAPI").Parse(yfr.Config))
	err := tmpl.Execute(configBuf, wd)
	yfr.Config = configBuf.String()
	return err
}

// InsertSteps ...
func InsertSteps(workloadDetails types.WorkloadDetails) {
	stageSteps := procureContent(workloadDetails)
	imageList := make(map[int]string)
	completedSteps := map[int]bool{}
	interval := 5
	tick := time.Tick(time.Duration(interval) * time.Second)
	timeout := time.Tick(20 * time.Minute)
	isEnd := false
	second := 0
	hasWorkloadFailed := false
	for (len(completedSteps) != len(stageSteps)) && !isEnd {
		select {
		case <-tick:
			if hasWorkloadFailed {
				log.Printf("Workload has failed. Exiting")
				// Shutdown all channels
				for id, singleDeleteChannel := range MapOfDeleteChannels {
					singleDeleteChannel <- types.WorkloadResult{
						ID:     id,
						Result: "Failed",
					}
					defer close(singleDeleteChannel)
					defer delete(MapOfDeleteChannels, id)
				}
				isEnd = true
			}
			second += interval
			log.Printf("\n\n%dth Second", second)
			log.Println(completedSteps)
			for index, singleStep := range stageSteps {
				log.Printf("%s - Checking Step. State = %s", singleStep.Task, singleStep.Status)
				// Check if each step if not executed, can be executed
				if (singleStep.Status != "Succeeded") && (singleStep.Status != "Triggered") && (singleStep.Status != "Failed") {
					log.Printf("%s - Step is not in Succeeded or Triggered or Failed State", singleStep.Task)
					// Check if the step is eligible for execution
					foundAnIncompleteRequiredStep := false
					log.Printf("%s - Checking if Step requirements are satisfied", singleStep.Task)
					if len(singleStep.Requires) <= len(completedSteps) {
						log.Printf("%s - Requires Steps count less than Completed Steps count", singleStep.Task)
						for _, singleRequiredStepID := range singleStep.Requires {
							if !completedSteps[singleRequiredStepID] {
								log.Printf("%s - Found an incomplete Step", singleStep.Task)
								foundAnIncompleteRequiredStep = true
								break
							}
						}
						if !foundAnIncompleteRequiredStep {
							workloadDetails.WorkloadID = workloadDetails.ID + "-" + strconv.Itoa(index)
							workloadDetails.StepID = index
							workloadDetails.Task = singleStep.Task
							workloadDetails.RegistryURL = "buildhub.myntra.com:5000"
							customPropertiesInBytes, err := json.Marshal(singleStep.Meta.CustomProperties)
							helpers.FailOnErr(err, nil)
							workloadDetails.CustomProperties = string(customPropertiesInBytes)
							if singleStep.Meta.Image != "" {
								if imageIndex, err := strconv.Atoi(singleStep.Meta.Image); err == nil {
									workloadDetails.Image = imageList[imageIndex]
									if workloadDetails.Image == "" {
										helpers.FailOnErr(fmt.Errorf("Trying to use an image which was not committed %s",
											singleStep.Task), nil)
									}
								} else {
									workloadDetails.Image = singleStep.Meta.Image
								}
							} else {
								log.Printf("Image not specified for step %s", singleStep.Task)
							}
							workloadDetails.CommitContainer = singleStep.CommitContainer
							// Trigger the API call to kuborch
							_, err = helpers.Post("http://localhost:5600/executeworkload", workloadDetails, nil)
							helpers.FailOnErr(err, nil)
							go func(index int, workloadDetails types.WorkloadDetails) {
								MapOfDeleteChannels[workloadDetails.WorkloadID] = make(chan types.WorkloadResult)
								// Hit kuborch API to create job
								everySecond := time.Tick(5 * time.Second)
								for {
									log.Printf("%s - Workload not complete", stageSteps[index].Task)
									select {
									case statusInChannel := <-MapOfDeleteChannels[workloadDetails.WorkloadID]:
										log.Printf("%s - Got a channel req - %v", stageSteps[index].Task, statusInChannel)
										if statusInChannel.Result == "Succeeded" {
											completedSteps[stageSteps[index].ID] = true
										} else {
											hasWorkloadFailed = true
											log.Printf("%s - Workload has failed. Stopping in 5 seconds", stageSteps[index].Task)
										}
										stageSteps[index].Status = statusInChannel.Result
										log.Printf("%s - Sleeping Done", stageSteps[index].Task)
										if workloadDetails.CommitContainer {
											// partner-terms-service-59-bf6532270193:clone
											imageList[index] = workloadDetails.Repo + "-" +
												strconv.Itoa(workloadDetails.PRID) + "-" +
												workloadDetails.SrcTopCommit + ":" +
												workloadDetails.Task
										}
										return
									case <-everySecond:
										log.Printf("%s - Nothing on the channels", stageSteps[index].Task)
										log.Printf("%s - Context - %s", stageSteps[index].Task, workloadDetails.WorkloadID)
									}
								}
							}(index, workloadDetails)
							stageSteps[index].Status = "Triggered"
							log.Printf("%s - Triggering Step", stageSteps[index].Task)
						} else {
							log.Printf("%s - Found an incomplete Required Step", singleStep.Task)
						}
					} else {
						log.Printf("%s - Step requirements NOT satisfied. Skipping", singleStep.Task)
					}
				} else {
					log.Printf("%s - Step State is %s. Skipping", singleStep.Task, singleStep.Status)
				}
			}
		case <-timeout:
			isEnd = true
			log.Println("Timed out")
		}
	}
}
