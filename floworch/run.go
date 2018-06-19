package main

import "github.com/myntra/shuttle-engine/types"

func updateStatus(run *types.Run, hasAnyWorkloadFailed bool) {
	if hasAnyWorkloadFailed {
		run.Status = types.FAILED
	} else {
		run.Status = types.SUCCEEDED
	}
	updateRunDetailsToDB(run)
}
