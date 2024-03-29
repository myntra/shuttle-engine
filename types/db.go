package types

import "time"

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
	ID                    string                  `json:"id" gorethink:"id" form:"id"`
	Stage                 string                  `json:"stage" gorethink:"stage" form:"stage"`
	Steps                 []Step                  `json:"steps" gorethink:"steps"`
	KVPairsSavedOnSuccess []KVPairsSavedOnSuccess `json:"kvPairsSavedOnSuccess" gorethink:"kvPairsSavedOnSuccess"`
	CandidateImage        string                  `json:"candidateImage,omitempty" gorethink:"candidateImage,omitempty"`
	Status                string                  `json:"status" gorethink:"status"`
	CreatedTime           time.Time               `json:"createdTime" gorethink:"createdTime"`
	UpdatedTime           time.Time               `json:"updatedTime" gorethink:"updatedTime"`
	StatusMessage         string                  `json:"statusMessage,omitempty" gorethink:"statusMessage,omitempty"`
}

// Step ...
type Step struct {
	ID                    int                     `yaml:"id" gorethink:"id"`
	Name                  string                  `yaml:"name" gorethink:"name"`
	IsCommand             bool                    `yaml:"isCommand" gorethink:"isCommand"`
	StepTemplate          string                  `yaml:"stepTemplate" gorethink:"stepTemplate"`
	Image                 string                  `yaml:"image" gorethink:"image"`
	K8SCluster            string                  `yaml:"k8scluster" gorethink:"k8sclustername"`
	ChartURL              string                  `json:"chartURL" gorethink:"chartURL"`
	ReleaseName           string                  `json:"releaseName" gorethink:"releaseName"`
	KubeConfig            string                  `json:"kubeConfig" gorethink:"kubeConfig"`
	Meta                  []Meta                  `yaml:"meta" gorethink:"meta"`
	Requires              []int                   `yaml:"requires" gorethink:"requires"`
	CommitContainer       bool                    `yaml:"commitContainer" gorethink:"commitContainer"`
	Status                string                  `yaml:"status" gorethink:"status"`
	StatusMessage         string                  `json:"statusMessage,omitempty" gorethink:"statusMessage,omitempty"`
	UniqueKey             string                  `yaml:"uniqueKey" gorethink:"uniqueKey"`
	Replacers             map[string]string       `yaml:"replacers" gorethink:"replacers"`
	IgnoreErrors          bool                    `yaml:"ignoreErrors" gorethink:"ignoreErrors"`
	IsNonCritical         bool                    `yaml:"isNonCritical" gorethink:"isNonCritical"`
	KVPairsSavedOnSuccess []KVPairsSavedOnSuccess `yaml:"kvPairsSavedOnSuccess" gorethink:"kvPairsSavedOnSuccess"`
	Duration              int                     `yaml:"int" gorethink:"duration"`
	Timeout               string                  `yaml:"timeout" gorethink:"timeout"`
	CandidateImage        string                  `json:"candidateImage,omitempty" gorethink:"candidateImage,omitempty"`
	Namespace             string                  `yaml:"namespace" gorethink:"namespace"`
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
