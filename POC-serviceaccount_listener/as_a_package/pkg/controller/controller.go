package controller

import (

	// "k8s.io/api/core/v1"

	"flag"
	"fmt"
	"log"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func GetKevents(kclient *kubernetes.Clientset) {
	clientset := kclient
	// vars
	initNamespace := "default"

	// clientset REQUIRED !
	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	// define namespace
	var ns string
	// if "--all-namespaces" then change "initNamespace" to empty string -> ""
	flag.StringVar(&ns, "namespace", initNamespace, "a namespace")
	flag.Parse()

	// display list of serviceaccounts in specific namespace
	dispSA, err := api.ServiceAccounts(initNamespace).List(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Initial data in namespace %s \n", initNamespace)
	for _, dispSA := range dispSA.Items {
		fmt.Printf("name: %s \n", dispSA.Name)
	}

	// enable watcher for serviceaccounts
	watcher, err := api.ServiceAccounts(initNamespace).Watch(listOptions)
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
