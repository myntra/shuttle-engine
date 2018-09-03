package main

import (
	"github.com/myntra/shuttle-engine/types"
	gorethink "gopkg.in/gorethink/gorethink.v4"
)

func getContent(flowOrchRequest types.FlowOrchRequest) (types.YAMLFromDB, error) {
	var yamlFromDB types.YAMLFromDB
	rdbSession, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
	})
	if err != nil {
		return yamlFromDB, err
	}
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
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
	})
	if err != nil {
		return run, err
	}
	_, err = gorethink.Table(run.Stage + "_runs").Filter(map[string]interface{}{
		"id": run.ID,
	}).Update(run).RunWrite(rdbSession)
	if err != nil {
		return run, err
	}
	if err != nil {
		return run, err
	}
	return run, nil
}
