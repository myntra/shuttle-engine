package main

import (
	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/types"
	r "gopkg.in/gorethink/gorethink.v4"
)

func getContent(flowOrchRequest types.FlowOrchRequest) (types.YAMLFromDB, error) {
	var yamlFromDB types.YAMLFromDB
	cursor, err := r.DB(config.GetConfig().ShuttleDBName).Table(flowOrchRequest.Stage + "_configs").Filter(map[string]interface{}{
		"id": flowOrchRequest.StageFilter,
	}).Run(config.RethinkSession)
	if err != nil {
		return yamlFromDB, err
	}
	defer cursor.Close()
	err = cursor.One(&yamlFromDB)
	if err != nil {
		return yamlFromDB, err
	}
	return yamlFromDB, nil
}

func updateRunDetailsToDB(run *types.Run) (*types.Run, error) {
	_, err := r.DB(config.GetConfig().ShuttleDBName).Table(run.Stage + "_runs").Filter(map[string]interface{}{
		"id": run.ID,
	}).Update(run).RunWrite(config.RethinkSession)
	if err != nil {
		return run, err
	}
	if err != nil {
		return run, err
	}
	return run, nil
}
