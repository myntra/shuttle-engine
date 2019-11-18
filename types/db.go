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

// Run ...
type Run struct {
	ID                    string                  `json:"id" gorethink:"id"`
	Stage                 string                  `json:"stage" gorethink:"stage"`
	Steps                 []Step                  `json:"steps" gorethink:"steps"`
	KVPairsSavedOnSuccess []KVPairsSavedOnSuccess `json:"kvPairsSavedOnSuccess" gorethink:"kvPairsSavedOnSuccess"`
	Status                string                  `json:"status" gorethink:"status"`
}

// Step ...
type Step struct {
	ID                    int                     `yaml:"id" gorethink:"id"`
	Name                  string                  `yaml:"name" gorethink:"name"`
	StepTemplate          string                  `yaml:"stepTemplate" gorethink:"stepTemplate"`
	Image                 string                  `yaml:"image" gorethink:"image"`
	Meta                  []Meta                  `yaml:"meta" gorethink:"meta"`
	Requires              []int                   `yaml:"requires" gorethink:"requires"`
	CommitContainer       bool                    `yaml:"commitContainer" gorethink:"commitContainer"`
	Status                string                  `yaml:"status" gorethink:"status"`
	UniqueKey             string                  `yaml:"uniqueKey" gorethink:"uniqueKey"`
	Replacers             map[string]string       `yaml:"replacers" gorethink:"replacers"`
	IgnoreErrors          bool                    `yaml:"ignoreErrors" gorethink:"ignoreErrors"`
	IsNonCritical         bool                    `yaml:"isNonCritical" gorethink:"isNonCritical"`
	KVPairsSavedOnSuccess []KVPairsSavedOnSuccess `yaml:"kvPairsSavedOnSuccess" gorethink:"kvPairsSavedOnSuccess"`
	Messages              []string                `yaml:"messages" gorethink:"mesages"` // Extra information added from runnung step by calling external api
}

// KVPairsSavedOnSuccess ...
type KVPairsSavedOnSuccess struct {
	Key   string `yaml:"key" gorethink:"key"`
	Value string `yaml:"value" gorethink:"value"`
}

// Meta ...
type Meta struct {
	Name           string      `yaml:"name" gorethink:"name"`
	Value          interface{} `yaml:"value" gorethink:"value"`
	ConvertedValue string      `yaml:"convertedValue" gorethink:"convertedValue"`
}
