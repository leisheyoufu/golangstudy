package client

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const timeLayoutStr = "2006-01-02 15:04:05" // time format must be this in golang

func GetNodesFromInformer(clientset *kubernetes.Clientset) {
	stopper := make(chan struct{})
	defer close(stopper)

	// init informer
	factory := informers.NewSharedInformerFactory(clientset, 0)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()
	defer runtime.HandleCrash()

	// start informer，list & watch
	go factory.Start(stopper)

	// sync resource
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	// customize handler
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: func(interface{}, interface{}) { fmt.Println("Informer event update not implemented") },
		DeleteFunc: func(interface{}) { fmt.Println("Informer event delete not implemented") },
	})

	// create lister
	nodeLister := nodeInformer.Lister()
	nodeList, err := nodeLister.List(labels.Everything())
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, node := range nodeList {
		fmt.Printf("Informer node %s\n", node.ObjectMeta.Name)
	}
}

func onAdd(obj interface{}) {
	node := obj.(*corev1.Node)
	fmt.Println("Informer event add a node:", node.Name)
}

func GetPodsFromInformer(clientset *kubernetes.Clientset) {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	stopper := make(chan struct{})
	defer close(stopper)
	go factory.Start(stopper)
	podInformer := factory.Core().V1().Pods()
	informer := podInformer.Informer()
	defer runtime.HandleCrash()
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	podLister := podInformer.Lister().Pods("kube-system")
	podList, err := podLister.List(labels.Everything())
	if err != nil {
		fmt.Println(err)
		return
	}
	var tempName string
	for _, pod := range podList {
		if len(tempName) == 0 {
			tempName = pod.ObjectMeta.Name
		}
		fmt.Printf("Informer list pod %s\n", pod.ObjectMeta.Name)
	}
	if len(tempName) == 0 {
		return
	}
	pod, err := podLister.Get(tempName)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Informer get pod name: %s, creation time: %s, spec node: %s, spec host name: %s， status pod ip: %s\n",
		pod.ObjectMeta.Name, pod.ObjectMeta.CreationTimestamp.Format(timeLayoutStr), pod.Spec.NodeName, pod.Spec.Hostname, pod.Status.PodIP)
}
