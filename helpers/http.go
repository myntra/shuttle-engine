package helpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ParseRequest Unmarshall `request` bodies into interface based `payload`
func ParseRequest(request *http.Request, payload interface{}) error {
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, payload)
	return err
}

// SendResponse ...
func SendResponse(response string, code int, w http.ResponseWriter) {
	inBytes, err := json.Marshal(map[string]string{
		"response": response,
	})
	PanicOnErrorAPI(err, w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(inBytes)
}
