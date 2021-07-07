package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
)

func GetPVC(index int, run *types.Run, flowOrchRequest types.FlowOrchRequest, completedSteps map[int]bool) (string, error) {

	if index == 0 {
		return createPVC(flowOrchRequest, 0, "-1", 0)
	} else {
		lastCompletedRequiredStep := -1
		for _, singleRequiredStepID := range run.Steps[index].Requires {
			lastCompletedRequiredStep = Max(singleRequiredStepID, lastCompletedRequiredStep)
		}
		if lastCompletedRequiredStep != index-1 && lastCompletedRequiredStep != -1 {
			log.Printf("fromPvcName %s", run.Steps[lastCompletedRequiredStep].Replacers["boundedPvcName"])
			return createPVC(flowOrchRequest, 1, run.Steps[lastCompletedRequiredStep].Replacers["boundedPvcName"], index)
		}

		for i := index - 1; i >= 0; i-- {
			if len(run.Steps[i].Replacers["boundedPvcName"]) != 0 {
				return run.Steps[i].Replacers["boundedPvcName"], nil
			}
		}
		return "NoPvcFound", errors.New("failed in getting any pvc")
	}

}

func createPVC(flowOrchRequest types.FlowOrchRequest, pvctype int, sourcePvc string, index int) (string, error) {

	logFile, err := os.OpenFile(flowOrchRequest.ID, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Unable to create a log file for the request.: %v", err)
		log.Print("Falling back to using current file")
	}
	defer logFile.Close()
	logger := log.New(logFile, flowOrchRequest.ID, log.Lshortfile)
	pvcStepName := "createPVC"
	pvcStepSteptemplate := "create-pvc"
	pvcStepUniqueKey := flowOrchRequest.Stage + "-" + flowOrchRequest.ID + "-" + strconv.Itoa(index) + "-pvc"

	if pvctype == 1 {
		pvcStepName += "Clone"
		pvcStepSteptemplate = "clone-pvc"
	}

	pvcStep := types.Step{
		ID:           pvctype,
		Name:         pvcStepName,
		StepTemplate: pvcStepSteptemplate,
		K8SCluster:   flowOrchRequest.K8SCluster,
		UniqueKey:    pvcStepUniqueKey,
		Replacers:    make(map[string]string),
	}
	pvcStep.Replacers["uniqueKey"] = pvcStepUniqueKey
	pvcStep.Replacers["name"] = pvcStep.Name
	pvcStep.Replacers["stage"] = flowOrchRequest.Stage
	pvcStep.Replacers["id"] = flowOrchRequest.ID
	pvcStep.Replacers["pvcName"] = pvcStepUniqueKey

	if pvctype == 1 {
		pvcStep.Replacers["fromPvcName"] = sourcePvc
	}
	_, err = helpers.Post("http://localhost:5600/executeworkload", pvcStep, nil)

	if err != nil {
		return "", err
	}

	c2 := make(chan types.WorkloadResult)

	go func() {
		deleteChannelDetails := types.DeleteChannelDetails{
			ID:            flowOrchRequest.ID,
			Stage:         flowOrchRequest.Stage,
			DeleteChannel: make(chan types.WorkloadResult),
			IgnoreErrors:  false,
			CreationTime:  time.Now(),
		}

		MapOfDeleteChannelDetails[pvcStep.UniqueKey] = &deleteChannelDetails
		logger.Printf("thread - %s - Started Delete Channel", pvcStep.Name)

		everySecond := time.Tick(5 * time.Second)
		for MapOfDeleteChannelDetails[pvcStep.UniqueKey] != nil {
			logger.Printf("thread - %s - Workload not complete", pvcStep.Name)
			select {
			case statusInChannel := <-MapOfDeleteChannelDetails[pvcStep.UniqueKey].DeleteChannel:
				logger.Printf("thread - %s - Got a channel req - %v", pvcStep.Name, statusInChannel)
				if statusInChannel.Result != types.FAILED && statusInChannel.Result != types.SUCCEEDED {
					logger.Printf("waiting for PVC bound , currently pvc in %s state", statusInChannel.Details)
					continue
				}
				close(MapOfDeleteChannelDetails[pvcStep.UniqueKey].DeleteChannel)
				delete(MapOfDeleteChannelDetails, pvcStep.UniqueKey)
				c2 <- statusInChannel

			case <-everySecond:
				logger.Printf("thread - %s - Nothing on the channels", pvcStep.Name)
			}
		}
	}()

	for {
		select {
		case res := <-c2:
			if res.Result == types.FAILED {
				return "", errors.New("failed in creating pvc")
			}
			return pvcStep.UniqueKey, nil
		case <-time.After(30 * time.Second):
			return "", errors.New("timeout: failed in creating pvc")
		}
	}

}

func Max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}
