package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	r "gopkg.in/gorethink/gorethink.v4"
	yaml "gopkg.in/yaml.v1"
)

//Port ...
var Port int

//RethinkSession ...
var RethinkSession *r.Session

// ConfigYamlFolder config file folder
var ConfigYamlFolder = "../"

// ConfigYaml config filename
var ConfigYaml = "config.yaml"

// Env Environment running in - dev or production. production is default
var Env = os.Getenv("ENV")

//Config ...
type Config struct {
	KuborchPort   int    `yaml:"kuborchPort"`
	FloworchPort  int    `yaml:"floworchPort"`
	BuildHubURL   string `yaml:"buildHubURL"`
	KuborchURL    string `yaml:"kuborchURL"`
	FloworchURL   string `yaml:"floworchURL"`
	KubConfigPath string `yaml:"kubConfigPath"`
	RethinkHost   string `yaml:"rethinkHost"`
	ShuttleDBName string `yaml:"shuttleDBName"`
}

var config Config

//InitFlags ...
func InitFlags() {
	flag.IntVar(&Port, "Port", 0, "Port On which Service Listens")
	flag.Parse()
}

//InitRethinkDBSession ...
func InitRethinkDBSession() error {
	log.Printf("InitRethinkDBSession:%s", config.RethinkHost)
	session, err := r.Connect(r.ConnectOpts{
		Address: config.RethinkHost,
		MaxIdle: 10,
		MaxOpen: 10,
	})

	if err != nil {
		log.Printf("Cannot connect to rethinkdb. Exiting...,Error: %s", err)
		return err
	}

	session.SetMaxOpenConns(10)

	RethinkSession = session

	return nil
}

// ReadConfig read configYaml
func ReadConfig() error {
	//unmarshal config.yaml
	log.Println("Env : " + Env)
	if Env == "dev" {
		ConfigYaml = "config_dev.yaml"
	}

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
