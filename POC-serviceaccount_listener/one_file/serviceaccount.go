package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getKevents() {
	// add to import -> metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// --------------
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	// --------------
	// https://medium.com/programming-kubernetes/building-stuff-with-the-kubernetes-api-part-4-using-go-b1d0e3c1c899
	// https://github.com/vladimirvivien/k8s-client-examples/blob/master/go/pvcwatch/main.go#L57

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

func main() {
	getKevents()
}
