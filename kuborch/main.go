package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/helpers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Clientset ...
var Clientset *kubernetes.Clientset

func main() {
	// load config.yml
	if err := config.ReadConfig(); err != nil {
		return
	}

	if err := config.InitRethinkDBSession(config.GetConfig().RethinkHost,
		config.GetConfig().RethinkDB); err != nil {
		return
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", config.GetConfig().KubConfigPath)
	helpers.FailOnErr(err)
	Clientset, err = kubernetes.NewForConfig(cfg)
	helpers.FailOnErr(err)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting up the server at %d", config.GetConfig().Port)
	router := mux.NewRouter()
	router.HandleFunc("/executeworkload", executeWorkload).Methods("Post")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.GetConfig().Port), router))
}
