package k8s

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getPodsStatus(kubeconfig, namespace string) (map[string]bool, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read kubeconfig %s err=%v\n", kubeconfig, err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init kubeconfig %s err=%v\n", kubeconfig, err)
		return nil, err
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get pod from k8s %s err=%v\n", kubeconfig, err)
		return nil, err
	}
	podStatus := make(map[string]bool)
	for _, pod := range pods.Items {
		podStatus[pod.ObjectMeta.Name] = true
		for _, container := range pod.Status.ContainerStatuses {
			if !container.Ready {
				podStatus[pod.ObjectMeta.Name] = false
			}
		}
	}
	return podStatus, nil
}
