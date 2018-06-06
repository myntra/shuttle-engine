package main

import (
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
	// Setup Stage Bucket
	// TODO

	// Convert Meta Tags
	for parser := 0; parser < len(stageSteps); parser++ {
		err = convertMetaTagsToReplacers(&stageSteps[parser], flowOrchRequest, parser)
		if err != nil {
			helpers.PanicOnErrorAPI(err, w)
			return
		}
	}
	go func() {
		// Start Ticker
		err = orchestrate(stageSteps, flowOrchRequest)
		if err != nil {
			helpers.FailOnErr(err, nil)
			return
		}
	}()
	helpers.SendResponse("Workload triggered", 200, w)
}
