package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Response ...
type Response struct {
	State string `json:"state"`
	Code  int    `json:"code"`
}

// FailOnErr Fail if error is not nil
func FailOnErr(err error, resChan *chan string) {
	if resChan != nil {
		*resChan <- err.Error()
	}
	if err != nil {
		PrintErr(err)
	}
}

// PrintErr ...
func PrintErr(err error) {
	fmt.Println(err.Error())
	debug.PrintStack()
	// log.Println(errors.Wrap(err, 3).ErrorStack())
}

// PanicOnErrorAPI ...
func PanicOnErrorAPI(err error, w http.ResponseWriter) {
	if err != nil {
		fmt.Println(err.Error())
		debug.PrintStack()
		// log.Println(errors.Wrap(err, 3).ErrorStack())
		SendResponse("Error : "+err.Error(), 500, w)
	}
}
