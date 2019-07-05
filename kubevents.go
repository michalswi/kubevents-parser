package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// namespace "default"
// go run kubevents.go --run-outside-k-cluster true

// random namespace
// go run kubevents.go --ns=mynamespace --run-outside-k-cluster true

// https://github.com/kubernetes/client-go
// https://medium.com/programming-kubernetes/building-stuff-with-the-kubernetes-api-part-4-using-go-b1d0e3c1c899

type eventsData struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timeup"`
}

const (
	apiVersion = "/api/v1"
)

var ServicePort = getEnv("SERVICEPORT", ":5000")
var countID int
var finalJson = make(map[string]interface{})
var datas []eventsData
var ns string
var initNs = getEnv("INITNAMESPACE", "default")

func main() {

	// TODO
	finalJson["status"] = "running"
	finalJson["error"] = "null"
	//
	var wg sync.WaitGroup
	wg.Add(2)
	go handleRequests(&wg)
	go getKevents(&wg)
	wg.Wait()
}

// get 'key' environment variable if exist on HOST machine otherwise return defalutValue
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

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

func getKubeconfig(runOutsideKcluster bool) (*kubernetes.Clientset, error) {

	// OPTION 1 - not in K8s
	// var kubeconfig *string
	// // if homeDir := homedir.HomeDir(); homeDir != "" {
	// homeDir := homedir.HomeDir()
	// if homeDir != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(homeDir, ".kube", "config"),
	// 		"(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	return nil, err
	// }
	// return kubernetes.NewForConfig(config)

	// OPTION 2 - works in K8s
	kubeConfigLocation := ""

	if runOutsideKcluster == true {
		homeDir := os.Getenv("HOME")
		kubeConfigLocation = filepath.Join(homeDir, ".kube", "config")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)

}

func getKevents(wg *sync.WaitGroup) {
	// add to import -> metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// add to import -> apiv1 "k8s.io/api/core/v1"

	t1 := time.Now()

	flag.StringVar(&ns, "ns", initNs, "a namespace")
	// flag.StringVar(&ns, "ns", "default", "a namespace")
	runOutsideKcluster := flag.Bool("run-outside-k-cluster", false, "Set this flag when running outside of the cluster.")
	flag.Parse()

	finalJson["namespace"] = ns

	// Create clientset to interact with the kubernetes cluster
	clientset, err := getKubeconfig(*runOutsideKcluster)
	if err != nil {
		log.Fatal(err)
	}
	// Print kubeconfig
	// fmt.Printf("%+v\n", clientset)

	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	// Display all events at once
	// getKevents := api.Events(ns)
	// listevent, err := getKevents.List(listOptions)
	// if err != nil {
	// 	panic(err)
	// }
	// for _, d := range listevent.Items {
	// 	// in 'k8s.io/api/core/v1/types.go' there is no 'Kind'
	// 	fmt.Printf("| name: %s | namespace: %s | reason: %s | message: %s |\n",
	// 		d.Name, d.Namespace, d.Reason, d.Message)
	// }

	// List Deployments
	// https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
	// deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	// fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	// list, err := deploymentsClient.List(metav1.ListOptions{})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, d := range list.Items {
	// 	fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	// }

	// Enable watcher for events
	watcher, err := api.Events(ns).Watch(listOptions)
	if err != nil {
		log.Printf("Verify provided namespace: %s", ns)
		// TODO, not really useful ouput: 'unknown (get events)'
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
