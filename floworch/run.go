package main

import "github.com/myntra/shuttle-engine/types"

func updateStatus(run *types.Run, workloadStatus string) {
	run.Status = workloadStatus
	updateRunDetailsToDB(run)
}
