package main

import (
	"flag"

	"github.com/leisheyoufu/golangstudy/k8sclient/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "config", "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	client.GetPodsFromClientset(clientset)
	client.GetNodesFromInformer(clientset)
	client.GetPodsFromInformer(clientset)
}
