package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/michalswi/kubevents-parser/POC/serviceaccount_listener/as_a_package/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

/*
> namespace "default"
go run main.go --run-outside-k-cluster true

> random namespace
go run main.go --ns=kube-system --run-outside-k-cluster true
*/

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

	// namespace
	var ns string
	flag.StringVar(&ns, "ns", "default", "Set this flag when changing default namespace.")

	// cloudconfig
	runOutsideKcluster := flag.Bool("run-outside-k-cluster", false, "Set this flag when running outside of the cluster.")
	flag.Parse()

	clientset, err := newClientSet(*runOutsideKcluster)
	if err != nil {
		log.Fatal(err)
	}

	controller.GetKevents(clientset, ns)
}
