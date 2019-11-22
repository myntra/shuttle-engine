package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/types"
)

// MapOfDeleteChannelDetails ...
var MapOfDeleteChannelDetails = make(map[string]types.DeleteChannelDetails)

// Metrics will have the value of an environment variable("METRICS") which enables the metrics if its value is "ON"
var Metrics string

func main() {
	router := mux.NewRouter()
	Metrics = os.Getenv("METRICS")
	if Metrics == "ON" {
		HealthCheckTelegraf()
	}

	router.HandleFunc("/execute", executeHandler).Methods("Post")
	router.HandleFunc("/callback", callbackHandler).Methods("Post")
	router.HandleFunc("/healthcheck", HealthCheckHandler).Methods("Get")
	router.HandleFunc("/queue", QueueStatusHandler).Methods("Get")
	port := 5500
	log.Printf("Starting server on :%d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
