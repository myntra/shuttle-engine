package main

import "github.com/myntra/shuttle-engine/types"

func saveKVPairs(step types.Step, run *types.Run) {
	if step.Status == types.SUCCEEDED {
		for _, singleKVPair := range step.KVPairsSavedOnSuccess {
			// Push values into DB
			run.KVPairsSavedOnSuccess = append(run.KVPairsSavedOnSuccess, singleKVPair)
		}
		updateRunDetailsToDB(run)
	}
}
