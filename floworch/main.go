package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/types"
)

// MapOfDeleteChannels ...
var MapOfDeleteChannels = make(map[string]chan types.WorkloadResult)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/execute", executeHandler).Methods("Post")
	router.HandleFunc("/callback", callbackHandler).Methods("Post")
	port := 5500
	log.Printf("Starting server on :%d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
