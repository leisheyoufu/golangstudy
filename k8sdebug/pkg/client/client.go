package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mattbaird/jsonpatch"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/watch"
	"k8s.io/kubernetes/pkg/client/conditions"
	// "k8s.io/kubernetes/pkg/client/conditions"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func GetClientset(config string) (*kubernetes.Clientset, *restclient.Config, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}
	return clientset, kubeConfig, nil
}

func GetPodsFromClientset(clientset *kubernetes.Clientset) {
	pods, err := clientset.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range pods.Items {
		fmt.Printf("Get pod from clientset %s\n", pod.ObjectMeta.Name)
	}
}

func GetDebugPod(clientset *kubernetes.Clientset, namespace string, podName string) (*v1.Pod, error) {
	if namespace == "" {
		namespace = "default"
	}
	pod, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	return pod, nil
}

func ForkDebugPod(clientset *kubernetes.Clientset, pod *v1.Pod) (*v1.Pod, error) {
	shareProcessNamespace := true
	directoryCreate := v1.HostPathDirectoryOrCreate
	copyPod := &v1.Pod{
		ObjectMeta: *pod.ObjectMeta.DeepCopy(),
		Spec:       *pod.Spec.DeepCopy(),
	}
	copyPod.ResourceVersion = ""
	copyPod.UID = ""
	copyPod.SelfLink = ""
	copyPod.CreationTimestamp = metav1.Time{}
	copyPod.OwnerReferences = []metav1.OwnerReference{}
	copyPod.Spec.ShareProcessNamespace = &shareProcessNamespace
	copyPod.Labels = map[string]string{
		"run": "debug-httpd",
	}
	copyPod.Name = fmt.Sprintf("debug-pod-%s", pod.Name)
	for _, container := range copyPod.Spec.Containers {
		container.LivenessProbe = nil
		container.ReadinessProbe = nil
	}
	copyPod.Spec.RestartPolicy = v1.RestartPolicyNever
	copyPod.Spec.Volumes = []v1.Volume{
		{
			Name: "docker",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/var/run/docker.sock",
				},
			},
		},
		{
			Name: "cgroup",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/sys/fs/cgroup",
				},
			},
		},
		{
			Name: "lxcfs",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/var/lib/lxc/lxcfs",
					Type: &directoryCreate,
				},
			},
		},
	}
	copyPod.Spec.Containers = append(copyPod.Spec.Containers, *newDebugContainer())
	pod, err := clientset.CoreV1().Pods(pod.Namespace).Create(copyPod)
	if err != nil {
		fmt.Printf("Error occurred while creating pod with debug container:  %v\n", err)
		return nil, err
	}
	watcher, err := clientset.CoreV1().Pods(pod.Namespace).Watch(metav1.SingleObject(pod.ObjectMeta))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fmt.Printf("Waiting for pod %s to run...\n", pod.Name)
	event, err := watch.UntilWithoutRetry(ctx, watcher, conditions.PodRunning)
	if err != nil {
		fmt.Printf("Error occurred while waiting for pod to run:  %v\n", err)
		return nil, err
	}
	pod = event.Object.(*v1.Pod)
	return pod, nil
}

func DeleteDebugPod(clientset *kubernetes.Clientset, namespace string, podName string) error {
	return clientset.CoreV1().Pods(namespace).Delete(podName, &metav1.DeleteOptions{})
}

// EphemeralContainer can not work
func UpdateDebugPod(clientset *kubernetes.Clientset, podName string, namespace string) (*v1.Pod, error) {
	pod, err := GetDebugPod(clientset, podName, namespace)
	if err != nil {
		return nil, err
	}
	oJson, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	containerName := pod.Spec.Containers[0].Name

	ephemeralCommon := v1.EphemeralContainerCommon{
		Name:            "busybox",
		Image:           "busybox:1.28.4",
		ImagePullPolicy: v1.PullIfNotPresent,
		Command: []string{
			"sleep",
			"3600",
		},
	}
	ephemeral := v1.EphemeralContainer{
		EphemeralContainerCommon: ephemeralCommon,
		TargetContainerName:      containerName,
	}
	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, ephemeral)
	mJson, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	patch, err := jsonpatch.CreatePatch(oJson, mJson)
	if err != nil {
		return nil, err
	}
	pb, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		return nil, err
	}
	fmt.Println(string(pb))
	attach := v1.EphemeralContainers{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "debug-patch",
			Namespace: "loch",
			Labels: map[string]string{
				"app": "patch-demo",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ephemeralcontainers",
		},
		EphemeralContainers: pod.Spec.EphemeralContainers,
	}
	e, err := clientset.CoreV1().Pods(namespace).UpdateEphemeralContainers(podName, &attach)
	//pod, err = clientset.CoreV1().Pods(namespace).Patch(podName, types.JSONPatchType, pb)
	if err != nil {
		return nil, err
	}
	fmt.Println(e)
	return pod, nil
	//return clientset.CoreV1().Pods(namespace).Update(pod)
}
