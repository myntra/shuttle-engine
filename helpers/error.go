package helpers

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response ...
type Response struct {
	State string `json:"state"`
	Code  int    `json:"code"`
}

// FailOnErr Fail if error is not nil
func FailOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// PanicOnErrorAPI ...
func PanicOnErrorAPI(err error, w http.ResponseWriter) {
	if err != nil {
		log.Println(err.Error())
		eRes := Response{
			State: "Error : " + err.Error(),
			Code:  500,
		}
		inBytes, _ := json.Marshal(eRes)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(inBytes)
	}
}
