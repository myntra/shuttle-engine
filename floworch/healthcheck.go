package main

import (
	"encoding/json"
	"net/http"

	"github.com/myntra/shuttle-engine/helpers"
)

// HealthCheckHandler ...
func HealthCheckHandler(w http.ResponseWriter, req *http.Request) {
	eRes := helpers.Response{
		State: "online",
		Code:  200,
	}
	inBytes, _ := json.Marshal(eRes)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}

// QueueStatusHandler ...
func QueueStatusHandler(w http.ResponseWriter, req *http.Request) {
	keys := make(map[string]string)
	for k, v := range MapOfDeleteChannelDetails {
		keys[k] = v.ID
	}

	response := make(map[string]interface{})

	response["jobs"] = len(keys)
	response["data"] = keys

	inBytes, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}
