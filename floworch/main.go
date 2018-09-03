package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/types"
)

// MapOfDeleteChannelDetails ...
var MapOfDeleteChannelDetails = make(map[string]types.DeleteChannelDetails)

func main() {
	// load  config.yml
	if err := config.ReadConfig(); err != nil {
		panic(err)
	}

	//RethinkDB Connection
	if err := config.InitRethinkDBSession(); err != nil {
		panic(err)
	}
	router := mux.NewRouter()
	router.HandleFunc("/execute", executeHandler).Methods("Post")
	router.HandleFunc("/callback", callbackHandler).Methods("Post")
	log.Printf("Starting server on :%d", config.GetConfig().FloworchPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.GetConfig().FloworchPort), router))
}
