package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"./pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// go run main.go --run-outside-k-cluster true

func newClientSet(runOutsideKcluster bool) (*kubernetes.Clientset, error) {

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

func main() {

	log.SetOutput(os.Stdout)

	runOutsideKcluster := flag.Bool("run-outside-k-cluster", false, "Set this flag when running outside of the cluster.")
	flag.Parse()

	clientset, err := newClientSet(*runOutsideKcluster)
	if err != nil {
		panic(err.Error())
	}

	controller.GetKevents(clientset)
}
