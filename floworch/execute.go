package main

import (
	"fmt"
	"net/http"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
)

func executeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse Request
	var flowOrchRequest types.FlowOrchRequest
	err := helpers.ParseRequest(r, &flowOrchRequest)
	if err != nil {
		helpers.PanicOnErrorAPI(err, w)
		return
	}

	// Get Content from DB
	yamlFromDB, err := getContent(flowOrchRequest)
	if err != nil {
		fmt.Println("Error in fetching YAML from DB for "+flowOrchRequest.ID, flowOrchRequest.Stage)
		helpers.PanicOnErrorAPI(err, w)
		return
	}

	// Replace Variables from API
	replaceFromAPI(&yamlFromDB, flowOrchRequest)

	// Extract Steps
	stageSteps, err := extractSteps(yamlFromDB)
	if err != nil {
		helpers.PanicOnErrorAPI(err, w)
		return
	}

	// Convert Meta Tags
	for parser := 0; parser < len(stageSteps); parser++ {
		err = convertMetaTagsToReplacers(&stageSteps[parser], flowOrchRequest, parser)
		// Setting all steps to QUEUED state
		stageSteps[parser].Status = types.QUEUED
		if err != nil {
			helpers.PanicOnErrorAPI(err, w)
			return
		}
	}

	// Start goroutine to complete API but run the workload
	go func() {
		run := &types.Run{
			ID:     flowOrchRequest.ID,
			Stage:  flowOrchRequest.Stage,
			Status: types.INPROGRESS,
			Steps:  stageSteps,
		}
		// Start Ticker
		workloadStatus := orchestrate(flowOrchRequest, run)
		// Update Run status
		updateStatus(run, workloadStatus)
	}()
	helpers.SendResponse("Workload triggered", 200, w)
}
