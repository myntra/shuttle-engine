package main

import (
	"encoding/json"
	"net/http"

	"github.com/myntra/shuttle-engine/helpers"
)

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
