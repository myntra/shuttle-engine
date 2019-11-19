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
		Address:  "dockinsrethink.myntra.com:28015",
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

// CopyAttributes ...
// Copies attributes from previosuly saved value
// - Messages
func CopyAttributes(run *types.Run) error {
	rdbSession, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
	})
	if err != nil {
		return err
	}
	defer rdbSession.Close()

	cursor, err := gorethink.Table(run.Stage + "_runs").
		Filter(map[string]interface{}{
			"id": run.ID,
		}).
		Run(rdbSession)

	var savedRun *types.Run
	err = cursor.One(&savedRun)

	for indx, val := range savedRun.Steps {
		run.Steps[indx].Messages = val.Messages
	}

	return nil
}
