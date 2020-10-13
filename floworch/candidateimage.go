package main

import "github.com/myntra/shuttle-engine/types"

func updateCandidateImage(Value string, run *types.Run) {
	run.CandidateImage = Value
	updateRunDetailsToDB(run)
}
