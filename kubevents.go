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
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
	// it's better to get port from ENVs - getEnv()
	// ServicePort = ":5000"
	apiVersion = "/api/v1"
)

var ServicePort = getEnv("SERVICEPORT", ":5000")
var countID int
var finalJson = make(map[string]interface{})
var datas []eventsData
var initNamespace = "default"

func main() {
	finalJson["status"] = "running"
	finalJson["error"] = "null"
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

	// // option 1 - TODO  - check if it works on k8s
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

	// option 2 - it works on k8s
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

	t1 := time.Now()

	// NOT NEEDED - moved to getKubeconfig() - instead I used 'kubeconfig related'
	// --------------------------------------
	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err)
	// }
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err)
	// }
	// --------------------------------------

	// kubeconfig related
	// TODO - move to separate function
	// --------------------------------------
	// Use ~/.kube/config rather than in cluster configuration
	// option 1 from getKubeconfig() DOESN'T care about 'runOutsideKcluster' !!
	// option 2 from getKubeconfig() REQUIRES 'go run kubevents --run-outside-k-cluster true'
	runOutsideKcluster := flag.Bool("run-outside-k-cluster", false, "Set this flag when running outside of the cluster.")
	flag.Parse()
	// Create clientset to interact with the kubernetes cluster
	clientset, err := getKubeconfig(*runOutsideKcluster)
	if err != nil {
		panic(err)
	}
	// print kubeconfig..
	// fmt.Printf("%+v\n", clientset)
	// --------------------------------------

	// define namespace
	var ns string
	// if "--all-namespaces" then change "initNamespace" to empty string -> ""
	flag.StringVar(&ns, "namespace", initNamespace, "a namespace")
	flag.Parse()

	// init, clientset REQUIRED!
	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	// display all events at once
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

	// enable watcher for events
	watcher, err := api.Events(ns).Watch(listOptions)
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

func getDeployments() {
	// add to import -> apiv1 "k8s.io/api/core/v1"

	// kubeconfig related
	// TODO - move to separate function
	// --------------------------------------
	// Use ~/.kube/config rather than in cluster configuration
	// option 1 from getKubeconfig() DOESN'T care about 'runOutsideKcluster' !!
	// option 2 from getKubeconfig() REQUIRES 'go run kubevents --run-outside-k-cluster 1'
	runOutsideKcluster := flag.Bool("run-outside-k-cluster", false, "Set this flag when running outside of the cluster.")
	flag.Parse()
	// Create clientset to interact with the kubernetes cluster
	clientset, err := getKubeconfig(*runOutsideKcluster)
	if err != nil {
		panic(err)
	}
	// --------------------------------------

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
