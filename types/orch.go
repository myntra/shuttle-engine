package types

import "time"

// WorkloadDetails ...
type WorkloadDetails struct {
	Stage            string `json:"stage"`
	Repo             string `json:"repo"`
	SrcBranch        string `json:"srcBranch"`
	DstBranch        string `json:"dstBranch"`
	PRID             int    `json:"prID"`
	SrcTopCommit     string `json:"sourceTopCommit"`
	ID               string `json:"id"`
	WorkloadID       string `json:"workloadID"`
	Task             string `json:"task"`
	RegistryURL      string `json:"registryURL"`
	Image            string `json:"image"`
	CommitContainer  bool   `json:"commitContainer"`
	RepoWebsite      string `json:"repoWebsite"`
	RepoSlug         string `json:"repoSlug"`
	CustomProperties string `yaml:"customProperties"`
	StepID           int    `json:"stepID"`
}

// WorkloadResult ...
type WorkloadResult struct {
	UniqueKey string `json:"uniqueKey"`
	Result    string `json:"result"`
	Details   string `json:"details"`
	Kind      string `json:"kind"`
}

// FlowOrchRequest ...
type FlowOrchRequest struct {
	Stage       string            `json:"stage"`
	StageFilter string            `json:"stageFilter"`
	Meta        map[string]string `json:"meta"`
	ID          string            `json:"id"`
	K8SCluster  string            `json:"k8scluster"`
}

// DeleteChannelDetails ...
type DeleteChannelDetails struct {
	ID            string              `json:"id"`
	Stage         string              `json:"stage"`
	DeleteChannel chan WorkloadResult `json:"deleteChannel"`
	IgnoreErrors  bool                `json:"ignoreErrors"`
	CreationTime  time.Time           `json:"creationTime"`
}

// Abort ...
// Used for {stage}_aborts
type Abort struct {
	ID          string `json:"id"`
	CreatedOn   string `json:"createdOn"`
	Description string `json:"description"`
}

const (
	// QUEUED ...
	QUEUED = "Queued"
	// INPROGRESS ...
	INPROGRESS = "In Progress"
	// SUCCEEDED ...
	SUCCEEDED = "Succeeded"
	// ABORTED ...
	ABORTED = "Aborted"
	// FAILED ...
	FAILED = "Failed"
)
