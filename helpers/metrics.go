package helpers

import (
	"bytes"
	// "github.com/myntra/shuttle-engine/config"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// TimeTracker : calculates the time taken by each step in any run(ci/crf)
func TimeTracker(start time.Time, requestType string, uniqueID string, stage string) {
	var serviceName string
	var pullRequestNumber string
	var commitID string
	var stepNumber string
	var runNumber string

	elapsed := time.Since(start).Milliseconds()

	if requestType == "crf" || requestType == "cf" {
		s := strings.Split(uniqueID, "-")
		l := len(s)

		/*
			Here we are slicing the uniqueID in reverse as the service name can or cannot contain hyphen in it
			this will handle both unique IDs "service-name-PR#-commitID-runId-stepNumber" or "servicename-PR#-commitID-runId-stepNumber"
		*/
		stepNumber = s[l-1]
		runNumber = s[l-2]
		commitID = s[l-3]
		pullRequestNumber = s[l-4]
		serviceName = map[bool]string{true: s[0], false: s[0] + "-" + s[1]}[l-5 == 0]
	}

	if os.Getenv("METRICS") == "ON" {
		data := `m_bizmetrics,app_name=floworch,request_type=` + requestType + `,service_name=` + serviceName + `,unique_id=` + uniqueID + `,pr_number=` + pullRequestNumber + `,commit_id=` + commitID + `,run_number=` + runNumber + `,step_number=` + stepNumber + `,stage=` + stage + ` duration=` + strconv.Itoa(int(elapsed))
		log.Printf("function:: %s, took:: %dms", uniqueID, elapsed)
		pushBusinessMetrics(data)

	} else {
		log.Printf("metrics is disabled")
	}

}

func pushBusinessMetrics(pushData string) {
	url := "http://localhost:8181/telegraf?precision=ms"

	data := []byte(pushData)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		log.Printf("response Status:%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body. ", err)
	}

	log.Printf("%s\n", body)
}
