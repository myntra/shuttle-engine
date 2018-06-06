package main

import (
	"fmt"
	"strings"

	"github.com/myntra/shuttle-engine/types"
)

func replaceFromAPI(yamlFromDB *types.YAMLFromDB, flowOrchRequest types.FlowOrchRequest) {
	for singleReplacer, value := range flowOrchRequest.Meta {
		yamlFromDB.Config = strings.Replace(yamlFromDB.Config, "{{."+singleReplacer+"}}", value, -1)
	}
	// Replace id
	yamlFromDB.Config = strings.Replace(yamlFromDB.Config, "{{.id}}", flowOrchRequest.ID, -1)
	fmt.Println(yamlFromDB.Config)
}
