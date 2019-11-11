package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/types"
)

// MapOfDeleteChannelDetails ...
var MapOfDeleteChannelDetails = make(map[string]types.DeleteChannelDetails)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/execute", executeHandler).Methods("Post")
	router.HandleFunc("/callback", callbackHandler).Methods("Post")
	router.HandleFunc("/healthcheck", HealthCheckHandler).Methods("Get")
	router.HandleFunc("/queue", QueueStatusHandler).Methods("Get")
	port := 5500
	log.Printf("Starting server on :%d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
