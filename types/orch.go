package types

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
	ID      string `json:"id"`
	Result  string `json:"result"`
	Details string `json:"details"`
}
