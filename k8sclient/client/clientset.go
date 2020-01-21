package client

import (
	"errors"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func GetPodsFromClientset(clientset *kubernetes.Clientset) {
	pods, err := clientset.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range pods.Items {
		fmt.Printf("Get pod from clientset %s\n", pod.ObjectMeta.Name)
	}
}

func GetServiceForDeployment(deployment string, namespace string, clientset *kubernetes.Clientset) (*corev1.Service, error) {
	svcs, err := clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	deploy, err := clientset.AppsV1().Deployments(namespace).Get(deployment, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, svc := range svcs.Items {
		if LabelContains(svc.Spec.Selector, deploy.Labels) {
			fmt.Fprintf(os.Stdout, "Find service %s for deployment %s\n", svc.Name, deployment)
			return &svc, nil
		}
	}
	fmt.Println("Could not find any services for deployment")
	return nil, errors.New("cannot find service for deployment")
}

func GetPodsForSvc(service string, namespace string, clientset *kubernetes.Clientset) (*corev1.PodList, error) {
	svc, err := clientset.CoreV1().Services(namespace).Get(service, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := clientset.CoreV1().Pods(namespace).List(listOptions)
	for _, pod := range pods.Items {
		fmt.Fprintf(os.Stdout, "Get pod name for service %s: %v\n", service, pod.Name)
	}
	return pods, err
}

func LabelContains(selector map[string]string, target map[string]string) bool {
	var ok bool
	var tVal string
	for sKey, sVal := range selector {
		if tVal, ok = target[sKey]; !ok {
			return false
		}
		if sVal != tVal {
			return false
		}
	}
	return true
}
