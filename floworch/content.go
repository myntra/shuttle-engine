package main

import (
	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/types"
	gorethink "gopkg.in/gorethink/gorethink.v4"
)

func getContent(flowOrchRequest types.FlowOrchRequest) (types.YAMLFromDB, error) {
	var yamlFromDB types.YAMLFromDB
	rdbSession, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  config.GetConfig().RethinkHost,
		Database: "shuttleservices",
	})
	if err != nil {
		return yamlFromDB, err
	}
	defer rdbSession.Close()
	cursor, err := gorethink.Table(flowOrchRequest.Stage + "_configs").Filter(map[string]interface{}{
		"id": flowOrchRequest.StageFilter,
	}).Run(rdbSession)
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
	rdbSession, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  config.GetConfig().RethinkHost,
		Database: "shuttleservices",
	})
	if err != nil {
		return run, err
	}
	defer rdbSession.Close()
	_, err = gorethink.Table(run.Stage+"_runs").Insert(run, gorethink.InsertOpts{
		Conflict: "update",
	}).RunWrite(rdbSession)
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
	rdbSession, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  config.GetConfig().RethinkHost,
		Database: "shuttleservices",
	})
	if err != nil {
		return abort, err
	}
	defer rdbSession.Close()
	cursor, err := gorethink.Table(stage + "_aborts").
		Filter(map[string]interface{}{
			"id": id,
		}).
		Run(rdbSession)
	err = cursor.One(&abort)

	return abort, err
}
