package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// https://github.com/kubernetes/client-go
// https://medium.com/programming-kubernetes/building-stuff-with-the-kubernetes-api-part-4-using-go-b1d0e3c1c899

type eventsData struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timeup"`
}

const (
	ServicePort = ":5000"
	apiVersion  = "/api/v1"
)

var countID int
var finalJson = make(map[string]interface{})
var datas []eventsData

func handleRequests(wg *sync.WaitGroup) {
	r := mux.NewRouter()
	myRouter := r.PathPrefix(apiVersion).Subrouter()
	myRouter.Path("/log").HandlerFunc(jsonToweb)
	fmt.Println("Start..")
	log.Fatal(http.ListenAndServe(ServicePort, myRouter))
	defer wg.Done()
}

func jsonToweb(w http.ResponseWriter, r *http.Request) {
	finalJson["data"] = datas
	json.NewEncoder(w).Encode(finalJson)
}

func passData(eName, eReason, eDiff string) {
	countID++
	datas = append(datas, eventsData{
		ID:        countID,
		Name:      eName,
		Reason:    eReason,
		Timestamp: eDiff,
	})
}

func getDeployments() {
	// add to import -> apiv1 "k8s.io/api/core/v1"

	// kubeconfig
	// 1, flags
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// 2, evn
	// kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// List Deployments
	// https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}

}

func getKevents(wg *sync.WaitGroup) {
	// add to import -> metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	t1 := time.Now()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	//1 code run outside of a cluster
	// clientcmd.BuildConfigFromFlags("", configFile)
	//2 code run in a cluster/the client code is destined to run in a pod
	// clientcmd.BuildConfigFromFlags("", "")
	//3 package rest to create the configuration from cluster information directly
	// rest.InClusterConfig()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// List Events
	var ns string
	// --all-namespaces -> ""
	flag.StringVar(&ns, "namespace", "default", "a namespace")
	flag.Parse()

	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	getKevents := api.Events(ns)

	// DISPLAY all events at once
	// listevent, err := getKevents.List(listOptions)
	// if err != nil {
	// 	panic(err)
	// }
	// for _, d := range listevent.Items {
	// 	// in 'k8s.io/api/core/v1/types.go' there is no 'Kind'
	// 	fmt.Printf("| name: %s | namespace: %s | reason: %s | message: %s |\n",
	// 		d.Name, d.Namespace, d.Reason, d.Message)
	// }

	// watcher
	watcher, err := getKevents.
		Watch(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	ch := watcher.ResultChan()

	// https://github.com/vladimirvivien/k8s-client-examples/blob/master/go/pvcwatch/main.go#L65
	for event := range ch {
		// add to import -> v1 "k8s.io/api/core/v1"
		ke, ok := event.Object.(*v1.Event)
		if !ok {
			log.Fatal("unexpected type")
		}
		switch event.Type {
		case watch.Added:
			log.Printf("Event added, name: %s, reason: %s, timestamp: %s\n", ke.Name, ke.Reason, ke.CreationTimestamp)
			t2 := ke.CreationTimestamp
			diff := t2.Sub(t1)
			// webserver considers only events which appeared after the script was run (diff > 0)
			if diff > 0 {
				sDiff := time.Time{}.Add(diff)
				passData(ke.Name, ke.Reason, sDiff.Format("15:04:05"))
			}

		case watch.Modified:
			log.Printf("Event modified, name: %s, reason: %s\n", ke.Name, ke.Reason)
		case watch.Deleted:
			log.Printf("Event deleted, name: %s, reason: %s\n", ke.Name, ke.Reason)
		case watch.Error:
			log.Printf("Event error, name: %s, reason: %s\n", ke.Name, ke.Reason)
		}
	}
	defer wg.Done()
}

func main() {
	finalJson["status"] = "running"
	finalJson["error"] = "null"
	var wg sync.WaitGroup
	wg.Add(2)
	go handleRequests(&wg)
	go getKevents(&wg)
	wg.Wait()
}
