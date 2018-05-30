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
}

// Meta ...
type Meta struct {
	Image            string                 `yaml:"image"`
	CustomProperties map[string]interface{} `yaml:"customProperties"`
}
