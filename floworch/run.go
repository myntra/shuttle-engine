package main

import (
	"math"
	"time"

	"github.com/myntra/shuttle-engine/types"
)

func updateStatus(run *types.Run, workloadStatus string) {
	run.Status = workloadStatus
	updateRunDetailsToDB(run)
}

func updateStepDuration(startTime time.Time, run *types.Run, stepIndex int) {
	seconds := int(math.Round(time.Since(startTime).Seconds())) // diff => round => convert to int
	run.Steps[stepIndex].Duration = seconds
}
