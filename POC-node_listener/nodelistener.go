package main

import (
	"flag"
	"log"
	"path/filepath"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// go get k8s.io/metrics/pkg/apis/metrics

func getKevents() {
	// add to import -> metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	// Display list of nodes
	// dispNode, err := api.Nodes().List(listOptions)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Initial nodes available: \n")
	// for _, no := range dispNode.Items {
	// 	fmt.Printf("name: %s \n", no.Name)
	// }

	// metrics
	// https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/metrics

	// Enable watcher
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
			// log.Printf("Event added, name: %s, kernel version: %v\n", ke.Name, ke.Status.NodeInfo.KernelVersion)
			// log.Printf("Event added, runtime version: %v\n", ke.Status.NodeInfo.ContainerRuntimeVersion)
			// log.Printf("Event added, volumes: %v\n", ke.Status.VolumesInUse)

			// TODO - get kubelet status
			// log.Printf("Event added, status: %v\n", ke.Status.Conditions)
			allCond := ke.Status.Conditions
			log.Printf("Event added, name: %s, status: %v\n", ke.Name, allCond[len(allCond)-1])

			// TODO - check statistics on node
			// https://stackoverflow.com/questions/52763291/get-current-resource-usage-of-a-pod-in-kubernetes-with-go-client
			// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/metrics/pkg/apis/metrics/v1beta1/types.go

			// case watch.Modified:
			// 	log.Printf("Event modified, name: %s, status: %v\n", ke.Name, ke.Status)
			// case watch.Deleted:
			// 	log.Printf("Event deleted, name: %s, labels: %s\n", ke.Name, ke.Labels)
			// case watch.Error:
			// 	log.Printf("Event error, name: %s, labels: %s\n", ke.Name, ke.Labels)
		}
	}
}

func main() {
	getKevents()
}
