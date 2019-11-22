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

// EnableMetrics will have the value of an environment variable("ENABLE_METRICS") which enables the metrics if its value is "ON"
var EnableMetrics, _ = strconv.ParseBool(os.Getenv("ENABLE_METRICS"))

func main() {
	router := mux.NewRouter()

	if EnableMetrics {
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
