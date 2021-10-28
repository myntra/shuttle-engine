package main

import (
	"log"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//CreateK8sClient ...
func CreateK8sClient(configPath string) (*kubernetes.Clientset, error) {

	cfg, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func CreateRestK8Client(configPath string) (*restclient.Config, *kubernetes.Clientset, error) {

	cfg, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Println("error in creating rest config", err)
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Println("error in creating clientset", err)
		return nil, nil, err
	}

	return cfg, clientset, err
}
