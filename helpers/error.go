package helpers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-errors/errors"
)

// Response ...
type Response struct {
	State string `json:"state"`
	Code  int    `json:"code"`
}

// FailOnErr Fail if error is not nil
func FailOnErr(err error) {
	if err != nil {
		log.Println(errors.Wrap(err, 3).ErrorStack())
	}
}

// PanicOnErrorAPI ...
func PanicOnErrorAPI(err error, w http.ResponseWriter) {
	if err != nil {
		log.Panic(err)
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
