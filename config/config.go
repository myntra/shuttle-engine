package config

import (
	yaml "gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
)

// Config ...
type Config struct {
	Filter         Filters `yaml:"filters"`
	TotalTimeTable string  `yaml:"totalTimeTable"`
	StepTimeTable  string  `yaml:"stepTimeTable"`
}

// Filters ...
type Filters map[string]string

var config Config

// ReadConfig read configYaml
func ReadConfig() error {

	config = Config{}
	configData, err := ioutil.ReadFile("/usr/bin/config.yaml")

	if err != nil {
		log.Println(err)
		return err
	}
	if err := yaml.Unmarshal([]byte(configData), &config); err != nil {
		log.Println(err)
	}

	return nil
}

// GetConfig ...
func GetConfig() Config {
	return config
}
