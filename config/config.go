package config

import (
	"fmt"
	yaml "gopkg.in/yaml.v1"
	"io/ioutil"
)

// ConfigYamlFolder config file folder
var ConfigYamlFolder = "../"

// ConfigYaml config filename
var ConfigYaml = "config.yaml"

//Config ...
type Config struct {
	Key1 string `yaml:"key1"`
	Key2 string `yaml:"key2"`
}

var config Config

// ReadConfig read configYaml
func ReadConfig() error {

	config = Config{}
	configData, err := ioutil.ReadFile(ConfigYamlFolder + ConfigYaml)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = yaml.Unmarshal([]byte(configData), &config)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// GetConfig ...
func GetConfig() Config {

	return config
}
