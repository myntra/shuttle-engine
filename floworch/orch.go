package main

import (
	"log"
	"strconv"
	"time"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
)

func orchestrate(flowOrchRequest types.FlowOrchRequest, run *types.Run) bool {
	updateRunDetailsToDB(run)
	imageList := make(map[int]string)
	completedSteps := map[int]bool{}
	interval := 5
	tick := time.Tick(time.Duration(interval) * time.Second)
	timeout := time.Tick(20 * time.Minute)
	isEnd := false
	second := 0
	hasWorkloadFailed := false
	for (len(completedSteps) != len(run.Steps)) && !isEnd {
		select {
		case <-tick:
			second += interval
			log.Printf("\n\n%dth Second", second)
			if hasWorkloadFailed {
				// TODO Is this needed?
				// hasWorkloadFailed = false
				log.Printf("A Workload has failed. Exiting/Aborting all other steps which do not ignoreErrors")
				// Shutdown all running channels for the uniqueKey
				isEnd = true
				for uniqueKey, singleDeleteChannelDetail := range MapOfDeleteChannelDetails {
					// If the uniqueKey matches and we cannot ignore errors for the workload
					if singleDeleteChannelDetail.ID == flowOrchRequest.ID {
						if !singleDeleteChannelDetail.IgnoreErrors {
							singleDeleteChannelDetail.DeleteChannel <- types.WorkloadResult{
								UniqueKey: uniqueKey,
								Result:    "Failed",
							}
							defer close(singleDeleteChannelDetail.DeleteChannel)
							defer delete(MapOfDeleteChannelDetails, uniqueKey)
							// TODO : Stop jobs if they are running
						} else {
							log.Printf("There are workloads which ignoreErrors. Running them")
							isEnd = false
						}
					}
				}
				// Change the state of all non-ignoreError steps to Aborted
				for index := 0; index < len(run.Steps); index++ {
					if run.Steps[index].Status == types.QUEUED {
						if !run.Steps[index].IgnoreErrors {
							run.Steps[index].Status = types.ABORTED
							completedSteps[run.Steps[index].ID] = true
						} else {
							log.Printf("There are workloads which ignoreErrors. Running them")
							isEnd = false
						}
					}
				}
			}
			log.Println(completedSteps)
			log.Println(imageList)
			wasThereAnAPIError := false
			for index := 0; (index < len(run.Steps)) && !wasThereAnAPIError; index++ {
				log.Printf("%s - Checking Step. State = %s", run.Steps[index].Name, run.Steps[index].Status)
				// Check if each step if not executed, can be executed
				if run.Steps[index].Status == types.QUEUED {
					log.Printf("%s - Step is Queued State", run.Steps[index].Name)
					// Check if the step is eligible for execution
					foundAnIncompleteRequiredStep := false
					log.Printf("%s - Checking if Step requirements are satisfied", run.Steps[index].Name)
					if len(run.Steps[index].Requires) <= len(completedSteps) {
						log.Printf("%s - Requires Steps count less than Completed Steps count", run.Steps[index].Name)
						for _, singleRequiredStepID := range run.Steps[index].Requires {
							if !completedSteps[singleRequiredStepID] {
								log.Printf("%s - Found an incomplete Step", run.Steps[index].Name)
								foundAnIncompleteRequiredStep = true
								break
							}
						}
						if !foundAnIncompleteRequiredStep {
							if run.Steps[index].Image != "" {
								if imageIndex, err := strconv.Atoi(run.Steps[index].Image); err == nil {
									run.Steps[index].Image = imageList[imageIndex]
								}
								run.Steps[index].Replacers["image"] = run.Steps[index].Image
							}
							_, err := helpers.Post("http://localhost:5600/executeworkload", run.Steps[index], nil)
							if err != nil {
								log.Printf("thread - %s - Workload API has failed. Stopping in 5 seconds", run.Steps[index].Name)
								hasWorkloadFailed = true
								wasThereAnAPIError = true
								break
							}
							go func(index int) {
								deleteChannelDetails := types.DeleteChannelDetails{
									ID:            flowOrchRequest.ID,
									Stage:         flowOrchRequest.Stage,
									DeleteChannel: make(chan types.WorkloadResult),
									IgnoreErrors:  run.Steps[index].IgnoreErrors,
								}
								MapOfDeleteChannelDetails[run.Steps[index].UniqueKey] = deleteChannelDetails
								log.Printf("thread - %s - Started Delete Channel", run.Steps[index].Name)
								log.Println(run.Steps[index].UniqueKey)
								log.Println(MapOfDeleteChannelDetails)
								// Hit kuborch API to create job
								everySecond := time.Tick(5 * time.Second)
								for {
									log.Printf("thread - %s - Workload not complete", run.Steps[index].Name)
									select {
									case statusInChannel := <-MapOfDeleteChannelDetails[run.Steps[index].UniqueKey].DeleteChannel:
										log.Printf("thread - %s - Got a channel req - %v", run.Steps[index].Name, statusInChannel)
										completedSteps[run.Steps[index].ID] = true
										if statusInChannel.Result != types.SUCCEEDED {
											hasWorkloadFailed = true
											log.Printf("thread - %s - Workload has failed. Stopping in 5 seconds", run.Steps[index].Name)
										}
										run.Steps[index].Status = statusInChannel.Result
										log.Printf("thread - %s - Sleeping Done", run.Steps[index].Name)
										if run.Steps[index].CommitContainer {
											imageList[index] = run.Steps[index].UniqueKey + ":" + run.Steps[index].Name
										}
										updateRunDetailsToDB(run)
										return
									// This might not be needed
									case <-everySecond:
										log.Printf("thread - %s - Nothing on the channels", run.Steps[index].Name)
										log.Printf("thread - %s - Context - %s", run.Steps[index].Name, run.Steps[index].UniqueKey)
									}
								}
							}(index)
							run.Steps[index].Status = types.TRIGGERED
							updateRunDetailsToDB(run)
							log.Printf("%s - Triggered Step", run.Steps[index].Name)
						} else {
							log.Printf("%s - Found an incomplete Required Step", run.Steps[index].Name)
						}
					} else {
						log.Printf("%s - Step requirements NOT satisfied. Skipping", run.Steps[index].Name)
					}
				} else {
					log.Printf("%s - Step State is %s. Skipping", run.Steps[index].Name, run.Steps[index].Status)
				}
			}
		case <-timeout:
			isEnd = true
			log.Println("Timed out")
		}
	}
	updateRunDetailsToDB(run)
	return hasWorkloadFailed
}
