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
			log.Println(imageList)
			for index := 0; index < len(stageSteps); index++ {
				log.Printf("%s - Checking Step. State = %s", stageSteps[index].Name, stageSteps[index].Status)
				// Check if each step if not executed, can be executed
				// if (singleStep.Status != "Succeeded") && (singleStep.Status != "Triggered") && (singleStep.Status != "Failed") {
				if stageSteps[index].Status == "" {
					log.Printf("%s - Step is not in Succeeded or Triggered or Failed State", stageSteps[index].Name)
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
								MapOfDeleteChannels[stageSteps[index].UniqueKey] = make(chan types.WorkloadResult)
								log.Printf("thread - %s - Started Delete Channel", stageSteps[index].Name)
								log.Println(stageSteps[index].UniqueKey)
								log.Println(MapOfDeleteChannels)
								// Hit kuborch API to create job
								everySecond := time.Tick(5 * time.Second)
								for {
									log.Printf("thread - %s - Workload not complete", stageSteps[index].Name)
									select {
									case statusInChannel := <-MapOfDeleteChannels[stageSteps[index].UniqueKey]:
										log.Printf("thread - %s - Got a channel req - %v", stageSteps[index].Name, statusInChannel)
										if statusInChannel.Result == "Succeeded" {
											completedSteps[stageSteps[index].ID] = true
										} else {
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
