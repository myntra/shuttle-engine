package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/helpers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Clientset ...
var Clientset *kubernetes.Clientset

// ConfigPath ...
var ConfigPath *string

func main() {
	ConfigPath = flag.String("configPath", "~/.kube/config", "Path to kube config")
	flag.Parse()
	cfg, err := clientcmd.BuildConfigFromFlags("", *ConfigPath)
	helpers.FailOnErr(err, nil)
	Clientset, err = kubernetes.NewForConfig(cfg)
	helpers.FailOnErr(err, nil)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	port := 5600
	log.Printf("Starting up the server at %d", port)
	router := mux.NewRouter()

	// intercepting http call and fetching/pushing metrics to localhost telegraf on 8181
	httpWrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mt := httpsnoop.CaptureMetrics(router, w, r)
		timeParser, _ := time.ParseDuration(mt.Duration.String())
		fmt.Printf("There are %.6f seconds in %v.\n", timeParser.Seconds(), timeParser)
		rt := fmt.Sprintf("%.6f", timeParser.Seconds()*1000)
		pushAppMetrics(r.URL.Path, r.Method, mt.Code, rt, mt.Written)
	})

	router.HandleFunc("/executeworkload", executeWorkload).Methods("Post")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), httpWrapper))
}

func pushAppMetrics(URL string, method string, statuscode int, responsetime string, writtenbytes int64) {

	log.Printf("Request metrics :::  method=%s, url=%s, status_code=%d, response_time=%s, written_bytes=%d", method, URL, statuscode, responsetime, writtenbytes)

	apiURL := "http://localhost:8181/telegraf?precision=ms"

	pushData := `m_appmetrics,app_name=ci_kuborch,url="` + URL + `",statuscode=` + strconv.Itoa(statuscode) + `,method="` + method + `" duration=` + responsetime
	log.Println(pushData)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(apiURL, "text/plain", bytes.NewBufferString(pushData))
	if err != nil {
		fmt.Println(err)
	}
	log.Printf(resp.Status)
}
