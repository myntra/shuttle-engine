package config

import (
	"flag"
	"log"

	r "gopkg.in/gorethink/gorethink.v4"
)

//Port ...
var Port int

//ShuttleRethinkSession ...
var ShuttleRethinkSession *r.Session

//BuildHub ...
var BuildHub string

//KuborchURL ...
var KuborchURL string

//FloworchURL ...
var FloworchURL string

//KubConfigPath ...
var KubConfigPath string

//InitFlags ...
func InitFlags() {
	flag.IntVar(&Port, "Port", 0, "Port On which Service Listens")
	flag.StringVar(&BuildHub, "BuildHub", "buildhub.myntra.com", "build hub")
	flag.StringVar(&KuborchURL, "KuborchURL", "kuborch.myntra.com", "KuborchURL")
	flag.StringVar(&FloworchURL, "FloworchURL", "floworch.myntra.com", "FloworchURL")
	flag.StringVar(&KubConfigPath, "KubConfigPath", "~/.kube/config", "Path to kube config")
	flag.Parse()
}

//InitShuttleRethinkDBSession ...
func InitShuttleRethinkDBSession() error {

	log.Printf("InitShuttleRethinkDBSession:dockinsrethink.myntra.com")
	session, err := r.Connect(r.ConnectOpts{
		Address:  "dockinsrethink.myntra.com:28015",
		Database: "shuttleservices",
		MaxIdle:  10,
		MaxOpen:  10,
	})

	if err != nil {
		log.Println("Cannot connect to rethinkdb. Exiting...")
		return err
	}

	session.SetMaxOpenConns(10)

	ShuttleRethinkSession = session

	return nil

}
