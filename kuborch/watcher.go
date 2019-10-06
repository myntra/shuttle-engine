package main

import (
	"fmt"
	"log"
	"time"

	"github.com/myntra/shuttle-engine/types"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	batchv1 "k8s.io/api/batch/v1"
)

// StatefulSetWatch ...
func StatefulSetWatch(clientset *kubernetes.Clientset, resultChan chan types.WorkloadResult, namespace string, listOpts metav1.ListOptions) {

	watcher, err := clientset.AppsV1().StatefulSets(namespace).Watch(listOpts)
	if err != nil {
		fmt.Println("Failure in creating the watcher")
		fmt.Println(err)
	}

	ch := watcher.ResultChan()
	defer watcher.Stop()
	for {
		select {
		case event := <-ch:
			if event.Type == watch.Deleted {
				resultChan <- types.WorkloadResult{
					Result:  types.SUCCEEDED,
					Details: "",
					Kind:    "StatefulSet",
				}
				return
			} else if event.Type == watch.Modified {
				sfs := event.Object.(*appsv1.StatefulSet)
				log.Printf("*sfs.Spec.Replicas=%d, sfs.Status=%+v", *sfs.Spec.Replicas, sfs.Status)
				if *sfs.Spec.Replicas == sfs.Status.Replicas &&
					sfs.Status.Replicas == sfs.Status.ReadyReplicas &&
					sfs.Status.Replicas == sfs.Status.UpdatedReplicas &&
					sfs.Status.CurrentReplicas == sfs.Status.UpdatedReplicas &&
					sfs.Status.CurrentRevision == sfs.Status.UpdateRevision {
					resultChan <- types.WorkloadResult{
						Result:  types.SUCCEEDED,
						Details: "",
						Kind:    "StatefulSet",
					}
					return
				}
			}
		case <-time.After(15 * time.Second):
			resultChan <- types.WorkloadResult{
				Result:  types.FAILED,
				Details: "Timed out while waiting for events with StatefulSet",
				Kind:    "StatefulSet",
			}
		}
	}
}

// ServiceWatch ...
func ServiceWatch(clientset *kubernetes.Clientset, resultChan chan types.WorkloadResult, namespace string, listOpts metav1.ListOptions) {

	watcher, err := clientset.CoreV1().Services(namespace).Watch(listOpts)
	if err != nil {
		fmt.Println("Failure in creating the watcher")
		fmt.Println(err)
	}

	ch := watcher.ResultChan()
	defer watcher.Stop()
	for {
		select {
		case event := <-ch:
			fmt.Println(event.Type)
			if event.Type == watch.Deleted || event.Type == watch.Added {
				resultChan <- types.WorkloadResult{
					Result:  types.SUCCEEDED,
					Details: "",
					Kind:    "Service",
				}
				return
			}
		case <-time.After(5 * time.Second):
			resultChan <- types.WorkloadResult{
				Result:  types.FAILED,
				Details: "Timed out while waiting for events with Service",
				Kind:    "StatefulSet",
			}
		}
	}
}

// JobWatch ...
func JobWatch(clientset *kubernetes.Clientset, resultChan chan types.WorkloadResult, namespace string, listOpts metav1.ListOptions) {
	watcher, err := clientset.BatchV1().Jobs(namespace).Watch(listOpts)
	if err != nil {
		log.Println("Failure in creating the watcher")
		log.Println(err)
		return
	}

	ch := watcher.ResultChan()
	defer watcher.Stop()
	for {
		select {
		case event := <-ch:
			if event.Object == nil {
				log.Println("-- Received nil event from closed channel, refreshing channel --")
				watcher, _ = clientset.BatchV1().Jobs(namespace).Watch(listOpts)
				ch = watcher.ResultChan()
				continue
			}
			job := event.Object.(*batchv1.Job)
			log.Printf("Job: %s -> Active: %d, Succeeded: %d, Failed: %d, Spec Completions: %d",
				job.Name, job.Status.Active, job.Status.Succeeded, job.Status.Failed, *job.Spec.Completions)
			switch event.Type {
			case watch.Modified:
				sendResponse := false
				log.Println("New modification poll")
				res := types.FAILED
				errMsg := "Job Failed on K8s"

				if job.Status.Failed > 0 {
					log.Println("Workload Failed")
					sendResponse = true
				}

				if job.Status.Active == 0 &&
					job.Status.Failed == 0 &&
					job.Status.Succeeded == *job.Spec.Completions {

					log.Println("Workload Succeeded")
					res = types.SUCCEEDED
					errMsg = ""
					sendResponse = true
				}

				if sendResponse {
					log.Println("Stopping Poll")
					resultChan <- types.WorkloadResult{
						// UniqueKey: uniqueKey,
						Result:  res,
						Details: errMsg,
						Kind:    "Job",
					}
					return
				}
			}
		case <-time.After(45 * time.Minute):
			log.Println("Timeout for Job !!")
			log.Println("Stopping Poll")
			resultChan <- types.WorkloadResult{
				// UniqueKey: uniqueKey,
				Result:  types.FAILED,
				Details: "timeout",
				Kind:    "Job",
			}
			return
		}
	}
}

// DeploymentWatch ...
func DeploymentWatch(clientset *kubernetes.Clientset, resultChan chan types.WorkloadResult, namespace string, listOpts metav1.ListOptions) {

	watcher, err := clientset.AppsV1().Deployments(namespace).Watch(listOpts)
	if err != nil {
		fmt.Println("Failure in creating the watcher")
		fmt.Println(err)
	}

	ch := watcher.ResultChan()
	defer watcher.Stop()
	for {
		select {
		case event := <-ch:
			if event.Type == watch.Deleted {
				resultChan <- types.WorkloadResult{
					Result:  types.SUCCEEDED,
					Details: "",
					Kind:    "Deployment",
				}
				return
			} else if event.Type == watch.Modified {
				dpl := event.Object.(*appsv1.Deployment)
				log.Printf("*dpl.Spec.Replicas=%d, dpl.Status=%+v", *dpl.Spec.Replicas, dpl.Status)
				if dpl.Status.UpdatedReplicas == *(dpl.Spec.Replicas) &&
					dpl.Status.Replicas == *(dpl.Spec.Replicas) &&
					dpl.Status.AvailableReplicas == *(dpl.Spec.Replicas) &&
					dpl.Status.ObservedGeneration >= dpl.Generation {
					resultChan <- types.WorkloadResult{
						Result:  types.SUCCEEDED,
						Details: "",
						Kind:    "Deployment",
					}
					return
				}
			}
		case <-time.After(180 * time.Second):
			resultChan <- types.WorkloadResult{
				Result:  types.FAILED,
				Details: "Timed out while waiting for events with Deployment",
				Kind:    "Deployment",
			}
		}
	}
}
