package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

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

	// init, clientset REQUIRED!
	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	// display list of nodes
	dispNode, err := api.Nodes().List(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Initial nodes available: \n")
	for _, no := range dispNode.Items {
		fmt.Printf("name: %s \n", no.Name)
	}

	// enable watcher
	watcher, err := api.Nodes().Watch(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	ch := watcher.ResultChan()

	for event := range ch {
		ke, ok := event.Object.(*v1.Node)
		if !ok {
			log.Fatal("unexpected type")
		}
		switch event.Type {
		case watch.Added:
			log.Printf("Event added, name: %s, kernel version: %v\n", ke.Name, ke.Status.NodeInfo.KernelVersion)
			log.Printf("Event added, name: %s, runtime version: %v\n", ke.Name, ke.Status.NodeInfo.ContainerRuntimeVersion)
			log.Printf("Event added, name: %s, volumes: %v\n", ke.Name, ke.Status.VolumesInUse)
			// TODO - get kubelet status
			log.Printf("Event added, name: %s, status: %v\n", ke.Name, ke.Status.Conditions)
		case watch.Modified:
			log.Printf("Event modified, name: %s, labels: %s\n", ke.Name, ke.Status)
		case watch.Deleted:
			log.Printf("Event deleted, name: %s, labels: %s\n", ke.Name, ke.Labels)
		case watch.Error:
			log.Printf("Event error, name: %s, labels: %s\n", ke.Name, ke.Labels)
		}
	}
}

func main() {
	getKevents()
}
