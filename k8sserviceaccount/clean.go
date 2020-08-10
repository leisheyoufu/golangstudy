package main

import (
	"fmt"
	"os"

	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getAutoTestNamespaces(clientset *kubernetes.Clientset) ([]string, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, namespace := range namespaces.Items {
		var val string
		var ok bool
		if val, ok = namespace.Labels["env"]; !ok {
			continue
		}
		if val != "autotest" {
			continue
		}
		if val, ok = namespace.Labels["app"]; !ok {
			continue
		}
		if val != "mysql" {
			continue
		}
		if val, ok = namespace.Labels["user"]; ok {
			fmt.Printf("Namespace %s: Marked by user, ignore.\n", namespace.Name)
			continue
		}
		if time.Now().Sub(namespace.CreationTimestamp.Time) < time.Duration(2)*time.Hour {
			fmt.Printf("Namespace %s: Created for less than 2 hours, ignore.\n", namespace.Name)
			continue
		}
		names = append(names, namespace.Name)
	}
	return names, nil
}

func deleteAutoTestNamespaces(clientset *kubernetes.Clientset, names []string) error {
	var err error
	for _, name := range names {
		err = clientset.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete namespace %s, Error: %+v\n", name, err)
		} else {
			fmt.Printf("Namespace %s deleted\n", name)
		}
	}
	return err
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %+v\n", err)
		os.Exit(1)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %+v\n", err)
		os.Exit(1)
	}
	names, err := getAutoTestNamespaces(clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %+v\n", err)
		os.Exit(1)
	}
	err = deleteAutoTestNamespaces(clientset, names)
	if err != nil {
		os.Exit(1)
	}
}
