package main

import "github.com/myntra/shuttle-engine/types"

func updateCandidateImage(Value string, run *types.Run, parameter string) {
	run.parameter = Value
	updateRunDetailsToDB(run)
}
