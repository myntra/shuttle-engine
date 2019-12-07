package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
)

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	workloadResult := types.WorkloadResult{}
	err := helpers.ParseRequest(r, &workloadResult)
	helpers.PanicOnErrorAPI(err, w)
	log.Println(workloadResult)
	if stepChannelDetails, isPresent := MapOfDeleteChannelDetails[workloadResult.UniqueKey]; isPresent {
		stepChannelDetails.DeleteChannel <- workloadResult
		log.Println("Sent channel status")
	} else {
		log.Println("Channel not found on process. There must have been a failed step which deleted this running step")
	}
	cr := types.CallbackResponse{
		State: "Callback ACK",
		Code:  200,
	}
	crInBytes, err := json.Marshal(&cr)
	helpers.PanicOnErrorAPI(err, w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(crInBytes)
}
