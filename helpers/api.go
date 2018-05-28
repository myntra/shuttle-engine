package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Post ...
func Post(url string, data interface{}, response interface{}) (int, error) {
	var payload *bytes.Buffer
	if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			return 500, err
		}
		payload = bytes.NewBuffer(d)
	}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return 500, err
	}
	req.Header.Add("content-type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	if res.StatusCode != 200 {
		return res.StatusCode, fmt.Errorf("GET %s - Bad status code: %d ", url, res.StatusCode)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 500, err
	}
	if response != nil {
		err = json.Unmarshal(body, response)
		if err != nil {
			return 500, err
		}
	}
	return 200, nil
}
