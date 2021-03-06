package helpers

import (
	"bytes"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/types"
)

// TimeTracker : calculates the time taken by each step in any run
func TimeTracker(enableMetrics bool, start time.Time, isTotalTimeMetrics bool, stage string, id string, stepName string, uniqueKey string, meta map[string]string) {
	if enableMetrics {
		var configData string
		var data string

		elapsed := time.Since(start).Milliseconds()
		filters := config.GetConfig().Filter

		for k, v := range filters {
			if stage == k {
				configData = k + "=" + meta[v] + ","
			}
		}
		if isTotalTimeMetrics {
			data = config.GetConfig().TotalTimeTable + `,app_name=floworch,stage=` + stage + `,` + configData + `unique_key=` + uniqueKey + ` duration=` + strconv.Itoa(int(elapsed))
		} else {
			data = config.GetConfig().StepTimeTable + `,app_name=floworch,stage=` + stage + `,step_id=` + id + `,step_name=` + stepName + `,` + configData + `unique_key=` + uniqueKey + ` duration=` + strconv.Itoa(int(elapsed))
		}
		pushBusinessMetrics(data)
	}
}

func pushBusinessMetrics(pushData string) {
	url := "http://localhost:8181/telegraf?precision=ms"
	data := []byte(pushData)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Println("Error reading request. ", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error reading response. ", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		log.Printf("pushData : %s", data)
		log.Printf("Response Status:%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body. ", err)
		log.Printf("%s\n", body)
	}
}

// UpdateStepInfo ...
func UpdateStepInfo(enableMetrics bool, startTime time.Time, isTotalTimeMetrics bool, floworchRequest types.FlowOrchRequest, run *types.Run, stepIndex int) {
	seconds := int(math.Round(time.Since(startTime).Seconds())) // diff => round => convert to int
	run.Steps[stepIndex].Duration = seconds

	step := run.Steps[stepIndex]
	TimeTracker(enableMetrics, startTime, isTotalTimeMetrics, floworchRequest.Stage, strconv.Itoa(step.ID), step.Name, step.UniqueKey, floworchRequest.Meta)
}
