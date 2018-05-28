package types

// YAMLFromRethink ...
type YAMLFromRethink struct {
	Name   string `json:"name"`
	Config string `json:"config"`
}

// Step ...
type Step struct {
	ID              int    `yaml:"id"`
	Task            string `yaml:"task"`
	Meta            Meta   `yaml:"meta"`
	CommitContainer bool   `yaml:"commitContainer"`
	Requires        []int  `yaml:"requires"`
	Status          string `yaml:"status"`
	// StepDetails     StepDetails `yaml:"stepDetails"`
}

// Meta ...
type Meta struct {
	Image string `json:"image"`
}
