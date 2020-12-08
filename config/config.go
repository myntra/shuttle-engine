package config

import (
	"io/ioutil"
	"log"

	gorethink "gopkg.in/gorethink/gorethink.v4"
	yaml "gopkg.in/yaml.v1"
)

// RethinkSession ...
var RethinkSession *gorethink.Session

// Config ...
type Config struct {
	RethinkHost    string  `yaml:"rethinkHost"`
	Filter         Filters `yaml:"filters"`
	TotalTimeTable string  `yaml:"totalTimeTable"`
	StepTimeTable  string  `yaml:"stepTimeTable"`
	ShuttleDBName  string  `yaml:"shuttleDBName"`
}

// Filters ...
type Filters map[string]string

var config Config

// ReadConfig read configYaml
func ReadConfig() error {

	config = Config{}
	configData, err := ioutil.ReadFile("../config.yaml")

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

// InitDatabaseSession ...
func InitDatabaseSession() error {
	log.Println("config.InitDatabaseSession:", config.RethinkHost)
	session, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  config.RethinkHost,
		Database: config.ShuttleDBName,
		MaxIdle:  10,
		MaxOpen:  10,
	})
	if err != nil {
		log.Println("Cannot connect to rethinkdb. Exiting...")
		return err
	}

	session.SetMaxOpenConns(10)

	RethinkSession = session

	return nil
}
