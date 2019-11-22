package main

import (
	"encoding/json"
	"net/http"

	"fmt"
	"github.com/myntra/shuttle-engine/helpers"
	"log"
	"net"
	"time"
)

// HealthCheckHandler ...
func HealthCheckHandler(w http.ResponseWriter, req *http.Request) {
	eRes := helpers.Response{
		State: "online",
		Code:  200,
	}
	inBytes, _ := json.Marshal(eRes)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}

// QueueStatusHandler ...
func QueueStatusHandler(w http.ResponseWriter, req *http.Request) {
	keys := make(map[string]string)
	for k, v := range MapOfDeleteChannelDetails {
		keys[k] = v.ID
	}

	response := make(map[string]interface{})

	response["jobs"] = len(keys)
	response["data"] = keys

	inBytes, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(inBytes)
}

// HealthCheckTelegraf : it will check telegraf's health on the local machine
/*
	To enable health check of telegraf, add this plugin [[outputs.health]] in telegraf.conf file
	[[outputs.health]]
  		service_address="http://localhost:8282"
		  read_timeout = "5s"
*/
func HealthCheckTelegraf() {
	addr := "localhost:8282"
	up := false
	for parser := 0; parser < 30 && !up; parser++ {
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second)
		} else {
			up = true
			defer conn.Close()
		}
	}
	if !up {
		panic(fmt.Errorf("%s never came up", addr))
	}
}
