package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/myntra/shuttle-engine/types"
)

// MapOfDeleteChannelDetails ...
var MapOfDeleteChannelDetails = make(map[string]types.DeleteChannelDetails)

func main() {
	router := mux.NewRouter()

	// intercepting http call and fetching/pushing metrics to localhost telegraf on 8181
	httpWrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mt := httpsnoop.CaptureMetrics(router, w, r)
		timeParser, _ := time.ParseDuration(mt.Duration.String())
		fmt.Printf("There are %.6f seconds in %v.\n", timeParser.Seconds(), timeParser)
		rt := fmt.Sprintf("%.6f", timeParser.Seconds()*1000)
		pushAppMetrics(r.URL.Path, r.Method, mt.Code, rt, mt.Written)
	})

	router.HandleFunc("/execute", executeHandler).Methods("Post")
	router.HandleFunc("/callback", callbackHandler).Methods("Post")
	port := 5500
	log.Printf("Starting server on :%d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), httpWrapper))
}

func pushAppMetrics(URL string, method string, statuscode int, responsetime string, writtenbytes int64) {

	log.Printf("Request metrics :::  method=%s, url=%s, status_code=%d, response_time=%s, written_bytes=%d", method, URL, statuscode, responsetime, writtenbytes)

	apiURL := "http://localhost:8181/telegraf?precision=ms"

	pushData := `m_appmetrics,app_name=ci_floworch,url="` + URL + `",statuscode=` + strconv.Itoa(statuscode) + `,method="` + method + `" duration=` + responsetime
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
