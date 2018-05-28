package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/types"
)

// MapOfDeleteChannels ...
var MapOfDeleteChannels = make(map[string]chan types.WorkloadResult)

func main() {
	// load config.yml
	if err := config.ReadConfig(); err != nil {
		return
	}

	if err := config.InitRethinkDBSession(config.GetConfig().RethinkHost,
		config.GetConfig().RethinkDB); err != nil {
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	router := mux.NewRouter()
	router.HandleFunc("/execute", executeHandler).Methods("Post")
	router.HandleFunc("/callback", callbackHandler).Methods("Post")
	log.Printf("Starting server on :%d", config.GetConfig().Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.GetConfig().Port), router))
}
