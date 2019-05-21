package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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

	// if empty string "" then "--all-namespaces"
	initNamespace := "default"

	// init, clientset REQUIRED!
	api := clientset.CoreV1()
	listOptions := metav1.ListOptions{}

	// Get all namespaces
	// allns, _ := clientset.Core().Namespaces().List(listOptions)
	// for _, d := range allns.Items {
	// 	fmt.Printf("name: %s \n", d.Name)
	// }

	// define namespace
	var ns string
	flag.StringVar(&ns, "namespace", initNamespace, "a namespace")
	flag.Parse()

	// display list of serviceaccounts in specific namespace
	// dispSA, err := api.ServiceAccounts(initNamespace).List(listOptions)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Initial data in namespace %s \n", initNamespace)
	// for _, sa := range dispSA.Items {
	// 	fmt.Printf("name: %s \n", sa.Name)
	// }

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
		case watch.Modified:
			log.Printf("Event modified, name: %s, secrets: %s in namespace: %s, age: %s\n", ke.Name, ke.Secrets, ke.Namespace, ke.CreationTimestamp)
			// TODO - ke.CreationTimestamp - get like 'Age' -> 2d23h
			// TODO - apply if '-prod' & '-test' for 'namespace-prod', 'namespace-test'
			// apply Role creation and RoleBinding to 'namespace-dev'
			if strings.Contains(ke.Name, "-dev") {
				setupRbac(clientset, ke.Name, ke.Namespace, "namespace-dev")
			}
			// TODO (not so important), move SA to 'namespace-dev' if role and binding created
			// api.ServiceAccounts(initNamespace).Update()

		case watch.Deleted:
			log.Printf("Event deleted, name: %s in namespace: %s\n", ke.Name, ke.Namespace)
			if strings.Contains(ke.Name, "-dev") {
				deleteRbac(clientset, ke.Name, ke.Namespace, "namespace-dev")
			}

		case watch.Error:
			log.Printf("Event error, name: %s in namespace: %s, age: %s\n", ke.Name, ke.Namespace, ke.CreationTimestamp)
		}
	}
}

func setupRbac(kclient *kubernetes.Clientset, saName string, saNamespace string, namespaceName string) {
	// add to import >> rbacv1 "k8s.io/api/rbac/v1"
	clientset := kclient
	roleName := fmt.Sprintf("%s-r", saName)
	roleBindName := fmt.Sprintf("%s-rb", saName)

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespaceName,
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"", "extensions", "apps"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"*"},
			},
		},
	}

	if _, err := clientset.RbacV1().Roles(namespaceName).Create(role); err != nil {
		log.Printf("Can't create role: %s, in namespace: %s", roleName, namespaceName)
		return
	}
	log.Printf("Role: %s, created in namespace: %s", roleName, namespaceName)

	roleBinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindName,
			Namespace: namespaceName,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      saName,
				Namespace: saNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	if _, err := clientset.RbacV1().RoleBindings(namespaceName).Create(roleBinding); err != nil {
		log.Printf("Can't create role binding: %s, in namespace: %s", roleBindName, namespaceName)
		return
	}
	log.Printf("Role binding: %s, created in namespace: %s", roleBindName, namespaceName)

	log.Printf("Service account added to default namespace.")
}

func deleteRbac(kclient *kubernetes.Clientset, saName string, saNamespace string, namespaceName string) {
	// add to import >> rbacv1 "k8s.io/api/rbac/v1"
	clientset := kclient
	roleName := fmt.Sprintf("%s-r", saName)
	roleBindName := fmt.Sprintf("%s-rb", saName)

	err := clientset.RbacV1().Roles(namespaceName).Delete(roleName, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Can't delete role: %s, from namespace: %s", roleName, namespaceName)
	}
	log.Printf("Role: %s, deleted from namespace: %s", roleName, namespaceName)

	// if err := clientset.RbacV1().RoleBindings(namespaceName).Delete(roleBindName, &metav1.DeleteOptions{}); err != nil {
	err = clientset.RbacV1().RoleBindings(namespaceName).Delete(roleBindName, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Can't delete role binding: %s, from namespace: %s", roleBindName, namespaceName)
	}
	log.Printf("Role binding: %s, deleted from namespace: %s", roleBindName, namespaceName)
}

func main() {
	getKevents()
}
