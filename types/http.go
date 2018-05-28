package types

// ExecuteResponse ...
type ExecuteResponse struct {
	State           string          `json:"state"`
	WorkloadDetails WorkloadDetails `json:"workloadDetails"`
	Code            int             `json:"code"`
}

// CallbackResponse ...
type CallbackResponse struct {
	State string `json:"state"`
	Code  int    `json:"code"`
}
