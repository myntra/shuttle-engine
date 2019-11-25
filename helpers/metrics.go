package helpers

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TimeTracker : calculates the time taken by each step in any run
func TimeTracker(enableMetrics bool, start time.Time, stage string, id string, stepTemplate string, uniqueKey string, stageFilter string) {
	if enableMetrics {
		elapsed := time.Since(start).Milliseconds()

		stageFilterList := strings.Split(stageFilter, "-")
		branch := stageFilterList[len(stageFilterList)-1]
		serviceName := strings.Replace(stageFilter, "-"+branch, "", -1)

		data := `m_bizmetrics,app_name=floworch,stage=` + stage + `,step_id=` + id + `,step_template=` + stepTemplate + `,unique_key=` + uniqueKey + `,stage_filter=` + stageFilter + `,branch=` + branch + `,service_name=` + serviceName + ` duration=` + strconv.Itoa(int(elapsed))

		log.Printf("StageFilter:: %s, Step:: %s took:: %dms", stageFilter, stepTemplate, elapsed)
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
		log.Printf("Response Status:%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body. ", err)
		log.Printf("%s\n", body)
	}
}
