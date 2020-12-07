package main

import (
	"time"

	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/types"
	gorethink "gopkg.in/gorethink/gorethink.v4"
)

func getContent(flowOrchRequest types.FlowOrchRequest) (types.YAMLFromDB, error) {
	var yamlFromDB types.YAMLFromDB
	cursor, err := gorethink.Table(flowOrchRequest.Stage + "_configs").Filter(map[string]interface{}{
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
	if run.CreatedTime.IsZero() {
		run.CreatedTime = time.Now()
	}

	run.UpdatedTime = time.Now()

	_, err = gorethink.Table(run.Stage+"_runs").Insert(run, gorethink.InsertOpts{
		Conflict: "update",
	}).RunWrite(config.RethinkSession)
	if err != nil {
		return run, err
	}
	if err != nil {
		return run, err
	}
	return run, nil
}

// GetAbortDetails ...
func GetAbortDetails(id string, stage string) (types.Abort, error) {

	var abort types.Abort
	cursor, err := gorethink.Table(stage + "_aborts").
		Filter(map[string]interface{}{
			"id": id,
		}).
		Run(config.RethinkSession)
	err = cursor.One(&abort)

	return abort, err
}
