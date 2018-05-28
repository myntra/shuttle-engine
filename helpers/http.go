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
