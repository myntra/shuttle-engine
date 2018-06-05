package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/myntra/shuttle-engine/types"
)

func convertMetaTagsToReplacers(step *types.Step, flowOrchRequest types.FlowOrchRequest, index int) error {
	// Initialize Replacers
	step.Replacers = make(map[string]string)
	// Database level meta tags being converted to Replacers
	for parser := 0; parser < len(step.Meta); parser++ {
		convertedValue := ""
		switch step.Meta[parser].Value.(type) {
		case string:
			twelveSpaces := "            "
			convertedValue = step.Meta[parser].Value.(string)
			if strings.Contains(convertedValue, "\n") {
				convertedValue = "|\n" + twelveSpaces + strings.Replace(convertedValue, "\n", "\n"+twelveSpaces, -1)
			}
			// convertedValue = strings.Replace(convertedValue, "'", "\'", -1)
		case map[string]interface{}:
			convertedValueInBytes, err := json.Marshal(step.Meta[parser].Value)
			if err != nil {
				return err
			}
			convertedValue = string(convertedValueInBytes)
		}
		// step.Meta[parser].ConvertedValue = convertedValue
		step.Replacers[step.Meta[parser].Name] = convertedValue
		fmt.Println(convertedValue)
	}

	// API level meta tags being converted to Replacers
	for metaKey, metaValue := range flowOrchRequest.Meta {
		step.Replacers[metaKey] = metaValue
	}
	step.Replacers["commitContainer"] = strconv.FormatBool(step.CommitContainer)
	step.Replacers["name"] = step.Name
	step.Replacers["stage"] = flowOrchRequest.Stage
	step.Replacers["id"] = flowOrchRequest.ID
	step.Replacers["uniqueKey"] = flowOrchRequest.Stage + "-" + flowOrchRequest.ID +
		"-" + strconv.Itoa(index)
	step.UniqueKey = flowOrchRequest.Stage + "-" + flowOrchRequest.ID +
		"-" + strconv.Itoa(index)
	return nil
}
