package types

// YAMLFromRethink ...
type YAMLFromRethink struct {
	Name   string `json:"name"`
	Config string `json:"config"`
}

// YAMLFromDB ...
type YAMLFromDB struct {
	ID     string `json:"id"`
	Config string `json:"config"`
}

// Step ...
type Step struct {
	ID              int               `yaml:"id"`
	Name            string            `yaml:"name"`
	StepTemplate    string            `yaml:"stepTemplate"`
	Image           string            `yaml:"image"`
	Meta            []Meta            `yaml:"meta"`
	Requires        []int             `yaml:"requires"`
	CommitContainer bool              `yaml:"commitContainer"`
	Status          string            `yaml:"status"`
	UniqueKey       string            `yaml:"uniqueKey"`
	Replacers       map[string]string `yaml:"replacers"`
}

// Meta ...
type Meta struct {
	Name           string      `yaml:"name"`
	Value          interface{} `yaml:"value"`
	ConvertedValue string      `yaml:"convertedValue"`

	// Value          interface{} `yaml:"value"`
	// InputType      string      `yaml:"inputType"`
}
