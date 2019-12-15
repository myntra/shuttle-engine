package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/myntra/shuttle-engine/config"
	"github.com/myntra/shuttle-engine/helpers"
	"github.com/myntra/shuttle-engine/types"
	gorethink "gopkg.in/gorethink/gorethink.v4"
)

// AbortRunHandler ...
func AbortRunHandler(w http.ResponseWriter, r *http.Request) {
	pathElements := strings.Split(r.URL.Path, "/")
	runID := pathElements[2]

	stageParam, ok := r.URL.Query()["stage"]
	if !ok || len(stageParam) < 1 {
		helpers.SendResponse(err.Error(), 500, w)
		return
	}

	stage := stageParam[0]

	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		helpers.SendResponse(err.Error(), 500, w)
		return
	}

	var abortrundetails types.Abort
	err = json.Unmarshal(data, &abortrundetails)
	if err != nil {
		helpers.SendResponse(err.Error(), 500, w)
		return
	}

	abortrundetails.ID = runID

	err = UpdateAbortStatus(abortrundetails, stage)

	if err != nil {
		helpers.SendResponse(err.Error(), 500, w)
		return
	}

	helpers.SendResponse("success", 200, w)
}

// UpdateAbortStatus ...
func UpdateAbortStatus(ard types.Abort, stage string) error {
	rdbSession, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  config.GetConfig().RethinkHost,
		Database: "shuttleservices",
	})
	if err != nil {
		return err
	}
	defer rdbSession.Close()

	_, err = gorethink.Table(stage+"_aborts").Insert(map[string]interface{}{
		"id":          ard.ID,
		"created_on":  time.Now().UTC().Format(time.ANSIC),
		"description": ard.Description,
	}, gorethink.InsertOpts{
		Conflict: "update",
	}).RunWrite(rdbSession)

	return err
}
