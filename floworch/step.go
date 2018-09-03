package main

import (
	"github.com/ghodss/yaml"
	"github.com/myntra/shuttle-engine/types"
)

func extractSteps(yamlFromDB types.YAMLFromDB) ([]types.Step, error) {
	var stageSteps []types.Step
	err := yaml.Unmarshal([]byte(yamlFromDB.Config), &stageSteps)
	return stageSteps, err
}
