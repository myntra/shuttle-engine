package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
)

func orchestrate(flowOrchRequest types.FlowOrchRequest, run *types.Run) string {
	defer helpers.TimeTracker(EnableMetrics, time.Now(), true, flowOrchRequest.Stage, "", "", flowOrchRequest.ID, flowOrchRequest.Meta)
	logFile, err := os.OpenFile(flowOrchRequest.ID, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Unable to create a log file for the request.: %v", err)
		log.Print("Falling back to using current file")
	}
	defer logFile.Close()
	logger := log.New(logFile, flowOrchRequest.ID, log.Lshortfile)
	updateRunDetailsToDB(run)
	imageList := make(map[int]string)
	completedSteps := map[int]bool{}
	interval := 5
	tick := time.Tick(time.Duration(interval) * time.Second)
	timeout := time.Tick(120 * time.Minute)
	isEnd := false
	second := 0
	hasWorkloadFailed := false
	isExternalAbort := false
	abortDescription := ""
	for (len(completedSteps) != len(run.Steps)) && !isEnd {
		select {
		case <-tick:
			// check if a run has been aborted
			val, err := GetAbortDetails(run.ID, run.Stage)
			if err != nil {
				logger.Printf("Error in fetching abort details for [%s] run %s\n", run.Stage, run.ID)
			} else {
				isExternalAbort = true
				hasWorkloadFailed = true
				abortDescription = val.Description
			}

			second += interval
			logger.Printf("\n\n%dth Second", second)
			if hasWorkloadFailed {
				// TODO Is this needed?
				// hasWorkloadFailed = false
				logger.Printf("A Workload has failed. Exiting/Aborting all other steps which do not ignoreErrors")
				// Shutdown all running channels for the uniqueKey
				isEnd = true
				for uniqueKey, singleDeleteChannelDetail := range MapOfDeleteChannelDetails {
					// If the uniqueKey matches and we cannot ignore errors for the workload
					if singleDeleteChannelDetail.ID == flowOrchRequest.ID {
						if !singleDeleteChannelDetail.IgnoreErrors {
							singleDeleteChannelDetail.DeleteChannel <- types.WorkloadResult{
								UniqueKey: uniqueKey,
								Result:    types.ABORTED,
							}
							// TODO : Hit kuborch API to abort executions
						} else {
							logger.Printf("There are workloads which ignoreErrors. Continuing with them")
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
							logger.Printf("There are workloads which ignoreErrors. Running them")
							isEnd = false
						}
					}
				}
			}
			logger.Println(completedSteps)
			logger.Println(imageList)
			wasThereAnAPIError := false
			for index := 0; (index < len(run.Steps)) && !wasThereAnAPIError; index++ {
				logger.Printf("%s - Checking Step. State = %s", run.Steps[index].Name, run.Steps[index].Status)
				// Check if each step if not executed, can be executed
				if run.Steps[index].Status == types.QUEUED {
					logger.Printf("%s - Step is Queued State", run.Steps[index].Name)
					// Check if the step is eligible for execution
					foundAnIncompleteRequiredStep := false
					logger.Printf("%s - Checking if Step requirements are satisfied", run.Steps[index].Name)
					if len(run.Steps[index].Requires) <= len(completedSteps) {
						logger.Printf("%s - Requires Steps count less than Completed Steps count", run.Steps[index].Name)
						for _, singleRequiredStepID := range run.Steps[index].Requires {
							if !completedSteps[singleRequiredStepID] {
								logger.Printf("%s - Found an incomplete Step", run.Steps[index].Name)
								foundAnIncompleteRequiredStep = true
								break
							}
						}
						if !foundAnIncompleteRequiredStep {
							if run.Steps[index].Image != "" {
								if imageIndex, err := strconv.Atoi(run.Steps[index].Image); err == nil {
									run.Steps[index].Image = imageList[imageIndex]
								}
								// Add image as a replacer since it is found
								run.Steps[index].Replacers["image"] = run.Steps[index].Image
							}
							// Add updated set of KV pairs as replacers
							for _, singleKVPair := range run.KVPairsSavedOnSuccess {
								run.Steps[index].Replacers[singleKVPair.Key] = singleKVPair.Value
							}

							if flowOrchRequest.K8SCluster == "" {
								run.Steps[index].K8SCluster = "default"
							} else {
								run.Steps[index].K8SCluster = flowOrchRequest.K8SCluster
							}

							if flowOrchRequest.ChartURL != "" {
								run.Steps[index].ChartURL = flowOrchRequest.ChartURL
							}

							if flowOrchRequest.Timeout != "" {
								run.Steps[index].Timeout = flowOrchRequest.Timeout
							}

							if flowOrchRequest.ReleaseName != "" {
								run.Steps[index].ReleaseName = flowOrchRequest.ReleaseName
							}

							if flowOrchRequest.KubeConfig != "" {
								run.Steps[index].KubeConfig = flowOrchRequest.KubeConfig
							}

							if flowOrchRequest.Namespace != "" {
								run.Steps[index].Namespace = flowOrchRequest.Namespace
							}

							// sending failure/abort info ENV variable
							run.Steps[index].Replacers["hasWorkloadFailed"] = strconv.FormatBool(hasWorkloadFailed)
							run.Steps[index].Replacers["isExternalAbort"] = strconv.FormatBool(isExternalAbort)
							run.Steps[index].Replacers["abortDescription"] = abortDescription

							// Channel creation before sending the request, to avoid fast responses causing
							// the race condition before channel is set
							deleteChannelDetails := types.DeleteChannelDetails{
								ID:            flowOrchRequest.ID,
								Stage:         flowOrchRequest.Stage,
								DeleteChannel: make(chan types.WorkloadResult),
								IgnoreErrors:  run.Steps[index].IgnoreErrors,
								CreationTime:  time.Now(),
							}
							MapOfDeleteChannelDetails[run.Steps[index].UniqueKey] = &deleteChannelDetails

							_, err := helpers.Post("http://localhost:5600/executeworkload", run.Steps[index], nil)
							if err != nil {
								logger.Printf("thread - %s - Workload API has failed. Stopping in 5 seconds", run.Steps[index].Name)
								hasWorkloadFailed = true
								wasThereAnAPIError = true
								close(MapOfDeleteChannelDetails[run.Steps[index].UniqueKey].DeleteChannel)
								delete(MapOfDeleteChannelDetails, run.Steps[index].UniqueKey)

								break
							}
							go func(index int) {
								defer helpers.UpdateStepInfo(EnableMetrics, time.Now(), false, flowOrchRequest, run, index)

								logger.Printf("thread - %s - Started Delete Channel", run.Steps[index].Name)
								logger.Println(run.Steps[index].UniqueKey)
								logger.Println(MapOfDeleteChannelDetails)
								// Hit kuborch API to create job
								everySecond := time.Tick(5 * time.Second)
								for MapOfDeleteChannelDetails[run.Steps[index].UniqueKey] != nil {
									logger.Printf("thread - %s - Workload not complete", run.Steps[index].Name)
									select {
									case statusInChannel := <-MapOfDeleteChannelDetails[run.Steps[index].UniqueKey].DeleteChannel:
										logger.Printf("thread - %s - Got a channel req - %v", run.Steps[index].Name, statusInChannel)
										completedSteps[run.Steps[index].ID] = true
										if statusInChannel.Result != types.SUCCEEDED {
											if run.Steps[index].IsNonCritical {
												logger.Printf("thread - %s - Workload has failed. But not critical to pipeline, Continuing ...", run.Steps[index].Name)
											} else {
												hasWorkloadFailed = true
												logger.Printf("thread - %s - Workload has failed. Stopping in 5 seconds", run.Steps[index].Name)
											}
										}

										if statusInChannel.Result == types.SUCCEEDED {
											if run.Steps[index].CandidateImage != "" {
												run.CandidateImage = run.Steps[index].CandidateImage
											}
										}
										run.Steps[index].Status = statusInChannel.Result
										logger.Printf("thread - %s - Sleeping Done", run.Steps[index].Name)
										if run.Steps[index].CommitContainer {
											imageList[index] = run.Steps[index].UniqueKey + ":" + run.Steps[index].Name
										}

										if len(statusInChannel.Details) > 0 {
											run.StatusMessage = run.StatusMessage + fmt.Sprintf("Step - %s, %s\n", run.Steps[index].Name, statusInChannel.Details)
											run.Steps[index].StatusMessage = fmt.Sprintf("%s", statusInChannel.Details)
										}
										saveKVPairs(run.Steps[index], run)
										close(MapOfDeleteChannelDetails[run.Steps[index].UniqueKey].DeleteChannel)
										delete(MapOfDeleteChannelDetails, run.Steps[index].UniqueKey)
									// This might not be needed
									case <-everySecond:
										logger.Printf("thread - %s - Nothing on the channels", run.Steps[index].Name)
										logger.Printf("thread - %s - Context - %s", run.Steps[index].Name, run.Steps[index].UniqueKey)
									}
								}
								log.Println("Exiting step callback select loop as channel is now nil", run.Steps[index].UniqueKey)
							}(index)
							run.Steps[index].Status = types.INPROGRESS
							updateRunDetailsToDB(run)
							logger.Printf("%s - Triggered Step", run.Steps[index].Name)
						} else {
							logger.Printf("%s - Found an incomplete Required Step", run.Steps[index].Name)
						}
					} else {
						logger.Printf("%s - Step requirements NOT satisfied. Skipping", run.Steps[index].Name)
					}
				} else {
					logger.Printf("%s - Step State is %s. Skipping", run.Steps[index].Name, run.Steps[index].Status)
				}
			}
		case <-timeout:
			isEnd = true
			hasWorkloadFailed = true
			logger.Println("Timed out")
		}
	}
	updateRunDetailsToDB(run)

	if hasWorkloadFailed {
		if isExternalAbort {
			return types.ABORTED
		}
		return types.FAILED
	}
	return types.SUCCEEDED

}
