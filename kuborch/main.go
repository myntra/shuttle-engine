package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/helpers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Clientset ...
var Clientset *kubernetes.Clientset

// ClientConfigMap ...
var ClientConfigMap map[string]ClientConfig

// ConfigPath ...
var ConfigPath *string

type configsList []string

func (i *configsList) String() string {
	return "Path to kube config"
}

func (i *configsList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var myConfigList configsList

//ClientConfig ...
type ClientConfig struct {
	Clientset  *kubernetes.Clientset
	ConfigPath string
}

func main() {
	if err := config.ReadConfig(); err != nil {
		log.Println(err)
		return
	}
	os.RemoveAll("./yaml")
	_ = os.Mkdir("./yaml", 0777)
	//Example for Running with configPath is ./kuborch -configPath=clustername:~/kube/config -configPath=clustername1:~/kube/config1
	flag.Var(&myConfigList, "configPath", "Please provide the Config Map as -configPath=<name>:<configPath>")
	flag.Parse()
	ClientConfigMap = make(map[string]ClientConfig)
	if len(myConfigList) == 0 {
		//if No configPath is provided then Below is the default Kube config Path
		defaultConfigPath := "~/.kube/config"
		cfg, err := clientcmd.BuildConfigFromFlags("", defaultConfigPath)
		helpers.FailOnErr(err, nil)
		Clientset, err = kubernetes.NewForConfig(cfg)
		helpers.FailOnErr(err, nil)
		ClientConfigMap["default"] = ClientConfig{Clientset: Clientset, ConfigPath: defaultConfigPath}
	} else {
		for _, singleConfigPath := range myConfigList {
			configPathSplit := strings.Split(singleConfigPath, ":")
			cfg, err := clientcmd.BuildConfigFromFlags("", configPathSplit[1])
			helpers.FailOnErr(err, nil)
			Clientset, err = kubernetes.NewForConfig(cfg)
			helpers.FailOnErr(err, nil)
			ClientConfigMap[configPathSplit[0]] = ClientConfig{Clientset: Clientset, ConfigPath: configPathSplit[1]}
		}
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	port := 5600
	log.Printf("Starting up the server at %d", port)
	router := mux.NewRouter()
	router.HandleFunc("/executeworkload", executeWorkload).Methods("Post")
	router.HandleFunc("/health", HealthCheckHandler).Methods("Get")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
