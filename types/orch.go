package types

// WorkloadDetails ...
type WorkloadDetails struct {
	ImageList       map[int]string `json:"imageList"`
	Stage           string         `json:"stage"`
	Repo            string         `json:"repo"`
	SrcBranch       string         `json:"srcBranch"`
	DstBranch       string         `json:"dstBranch"`
	PRID            int            `json:"prID"`
	SrcTopCommmit   string         `json:"sourceTopCommit"`
	ID              string         `json:"id"`
	WorkloadID      string         `json:"workloadID"`
	Task            string         `json:"task"`
	RegistryURL     string         `json:"registryURL"`
	Image           string         `json:"image"`
	CommitContainer bool           `json:"commitContainer"`
	RepoWebsite     string         `json:"repoWebsite"`
	RepoSlug        string         `json:"repoSlug"`
}

// WorkloadResult ...
type WorkloadResult struct {
	ID     string `json:"id"`
	Result string `json:"result"`
}
