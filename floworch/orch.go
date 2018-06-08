package main

import (
	"log"
	"strconv"
	"time"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
)

func orchestrate(stageSteps []types.Step, flowOrchRequest types.FlowOrchRequest) error {
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
				for index := 0; index < len(stageSteps); index++ {
					if stageSteps[index].Status == "" {
						if !stageSteps[index].IgnoreErrors {
							stageSteps[index].Status = "Aborted"
							completedSteps[stageSteps[index].ID] = true
						} else {
							log.Printf("There are workloads which ignoreErrors. Running them")
							isEnd = false
						}
					}
				}
			}
			log.Println(completedSteps)
			log.Println(imageList)
			for index := 0; index < len(stageSteps); index++ {
				log.Printf("%s - Checking Step. State = %s", stageSteps[index].Name, stageSteps[index].Status)
				// Check if each step if not executed, can be executed
				if stageSteps[index].Status == "" {
					log.Printf("%s - Step is not in Succeeded or Triggered or Failed or Aborted State", stageSteps[index].Name)
					// Check if the step is eligible for execution
					foundAnIncompleteRequiredStep := false
					log.Printf("%s - Checking if Step requirements are satisfied", stageSteps[index].Name)
					if len(stageSteps[index].Requires) <= len(completedSteps) {
						log.Printf("%s - Requires Steps count less than Completed Steps count", stageSteps[index].Name)
						for _, singleRequiredStepID := range stageSteps[index].Requires {
							if !completedSteps[singleRequiredStepID] {
								log.Printf("%s - Found an incomplete Step", stageSteps[index].Name)
								foundAnIncompleteRequiredStep = true
								break
							}
						}
						if !foundAnIncompleteRequiredStep {
							if stageSteps[index].Image != "" {
								if imageIndex, err := strconv.Atoi(stageSteps[index].Image); err == nil {
									stageSteps[index].Image = imageList[imageIndex]
								}
								stageSteps[index].Replacers["image"] = stageSteps[index].Image
							}
							_, err := helpers.Post("http://localhost:5600/executeworkload", stageSteps[index], nil)
							if err != nil {
								return err
							}
							go func(index int) {
								deleteChannelDetails := types.DeleteChannelDetails{
									ID:            flowOrchRequest.ID,
									Stage:         flowOrchRequest.Stage,
									DeleteChannel: make(chan types.WorkloadResult),
									IgnoreErrors:  stageSteps[index].IgnoreErrors,
								}
								MapOfDeleteChannelDetails[stageSteps[index].UniqueKey] = deleteChannelDetails
								log.Printf("thread - %s - Started Delete Channel", stageSteps[index].Name)
								log.Println(stageSteps[index].UniqueKey)
								log.Println(MapOfDeleteChannelDetails)
								// Hit kuborch API to create job
								everySecond := time.Tick(5 * time.Second)
								for {
									log.Printf("thread - %s - Workload not complete", stageSteps[index].Name)
									select {
									case statusInChannel := <-MapOfDeleteChannelDetails[stageSteps[index].UniqueKey].DeleteChannel:
										log.Printf("thread - %s - Got a channel req - %v", stageSteps[index].Name, statusInChannel)
										completedSteps[stageSteps[index].ID] = true
										if statusInChannel.Result != "Succeeded" {
											hasWorkloadFailed = true
											log.Printf("thread - %s - Workload has failed. Stopping in 5 seconds", stageSteps[index].Name)
										}
										stageSteps[index].Status = statusInChannel.Result
										log.Printf("thread - %s - Sleeping Done", stageSteps[index].Name)
										if stageSteps[index].CommitContainer {
											imageList[index] = stageSteps[index].UniqueKey + ":" + stageSteps[index].Name
										}
										return
									case <-everySecond:
										log.Printf("thread - %s - Nothing on the channels", stageSteps[index].Name)
										log.Printf("thread - %s - Context - %s", stageSteps[index].Name, stageSteps[index].UniqueKey)
									}
								}
							}(index)
							stageSteps[index].Status = "Triggered"
							log.Printf("%s - Triggered Step", stageSteps[index].Name)
						} else {
							log.Printf("%s - Found an incomplete Required Step", stageSteps[index].Name)
						}
					} else {
						log.Printf("%s - Step requirements NOT satisfied. Skipping", stageSteps[index].Name)
					}
				} else {
					log.Printf("%s - Step State is %s. Skipping", stageSteps[index].Name, stageSteps[index].Status)
				}
			}
		case <-timeout:
			isEnd = true
			log.Println("Timed out")
		}
	}
	return nil
}
