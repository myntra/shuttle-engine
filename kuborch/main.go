package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/helpers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Clientset ...
var Clientset *kubernetes.Clientset

var ConfigPath *string

func main() {
	ConfigPath = flag.String("configPath", "~/.kube/config", "Path to kube config")
	flag.Parse()
	cfg, err := clientcmd.BuildConfigFromFlags("", *ConfigPath)
	helpers.FailOnErr(err)
	Clientset, err = kubernetes.NewForConfig(cfg)
	helpers.FailOnErr(err)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	port := 5600
	log.Printf("Starting up the server at %d", port)
	router := mux.NewRouter()
	router.HandleFunc("/executeworkload", executeWorkload).Methods("Post")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
