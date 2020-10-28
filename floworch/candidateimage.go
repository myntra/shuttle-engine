package main

import "github.com/myntra/shuttle-engine/types"

func updateKey(run *types.Run, parameter string, value string) {
	run.parameter = value
	updateRunDetailsToDB(run)
}
