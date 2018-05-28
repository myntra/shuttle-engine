package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
)

// MapOfDeleteChannels ...
var MapOfDeleteChannels = make(map[string]chan types.WorkloadResult)

func main() {
	config.InitFlags()

	err := config.InitShuttleRethinkDBSession()
	if err != nil {
		helpers.FailOnErr(err)
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	router := mux.NewRouter()
	router.HandleFunc("/execute", executeHandler).Methods("Post")
	router.HandleFunc("/callback", callbackHandler).Methods("Post")
	log.Printf("Starting server on :%d", config.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), router))
}
