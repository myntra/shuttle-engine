package main

import "github.com/myntra/shuttle-engine/types"

func updateStatus(run *types.Run, workloadStatus string) {
	// if hasAnyWorkloadFailed {
	// 	run.Status = types.FAILED
	// } else {
	// 	run.Status = types.SUCCEEDED
	// }
	run.Status = workloadStatus
	updateRunDetailsToDB(run)
}
