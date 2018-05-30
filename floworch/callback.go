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
	if stepChannel, isPresent := MapOfDeleteChannels[workloadResult.ID]; isPresent {
		stepChannel <- workloadResult
		defer close(stepChannel)
		defer delete(MapOfDeleteChannels, workloadResult.ID)
		log.Println("Sent channel status")
	} else {
		log.Println("Channel not found on process. Should send to raft here")
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
