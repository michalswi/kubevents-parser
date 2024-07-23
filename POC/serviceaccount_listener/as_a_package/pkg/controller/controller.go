package controller

import (

	// "k8s.io/api/core/v1"

	"context"
	"fmt"
	"log"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func GetKevents(kclient *kubernetes.Clientset, ns string) {

	clientset := kclient

	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	// Display list of serviceaccounts in specific namespace
	dispSA, err := api.ServiceAccounts(ns).List(context.TODO(), listOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Initial data in namespace: %s \n", ns)
	for _, dispSA := range dispSA.Items {
		fmt.Printf("name: %s \n", dispSA.Name)
	}

	// Enable watcher for serviceaccounts
	watcher, err := api.ServiceAccounts(ns).Watch(context.TODO(), listOptions)
	if err != nil {
		log.Fatal(err)
	}
	ch := watcher.ResultChan()

	for event := range ch {
		ke, ok := event.Object.(*v1.ServiceAccount)
		if !ok {
			log.Fatal("unexpected type")
		}
		switch event.Type {
		case watch.Added:
			log.Printf("Event added, name: %s, secrets: %s\n", ke.Name, ke.Secrets)
			// TODO - ke.CreationTimestamp - get like 'Age' -> 2d23h
			if strings.Contains(ke.Name, "-dev") {
				// TODO - apply RoleBinding to 'namespace-dev'
				log.Printf("It works!")
			}
			// TODO - apply RoleBinding if '-prod' & '-test' for 'namespace-prod', 'namespace-test'
		case watch.Modified:
			log.Printf("Event modified, name: %s, secrets: %s, age: %s\n", ke.Name, ke.Secrets, ke.CreationTimestamp)
		case watch.Deleted:
			log.Printf("Event deleted, name: %s, secrets: %s, age: %s\n", ke.Name, ke.Secrets, ke.CreationTimestamp)
		case watch.Error:
			log.Printf("Event error, name: %s, secrets: %s, age: %s\n", ke.Name, ke.Secrets, ke.CreationTimestamp)
		}
	}
}
