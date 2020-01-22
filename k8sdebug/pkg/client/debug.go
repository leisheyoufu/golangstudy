package client

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/exec"
)

func Debug(kubeCli *kubernetes.Clientset, config *restclient.Config, streamOption exec.StreamOptions, namespace string, podName string) error {
	pod, err := GetDebugPod(kubeCli, namespace, podName)
	if err != nil {
		return err
	}

	forkPod, err := ForkDebugPod(kubeCli, pod)
	if err != nil {
		return err
	}
	defer DeleteDebugPod(kubeCli, namespace, forkPod.Name)
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go SignalExit(c, kubeCli, namespace, forkPod.Name)
	return RemoteCommand(kubeCli, config, streamOption, namespace, forkPod.Name, "debug-bash")
}

func RemoteCommand(client kubernetes.Interface, config *restclient.Config, streamOption exec.StreamOptions, namespace string, podName string,
	container string) error {
	t := streamOption.SetupTTY()
	fn := func() error {
		cmd := []string{
			"bash",
		}

		req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
			Namespace(namespace).SubResource("exec")
		if container != "" {
			req = req.Param("container", container)
		}
		option := &v1.PodExecOptions{
			Container: container,
			Command:   cmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    false,
			TTY:       t.Raw,
		}
		req.VersionedParams(
			option,
			scheme.ParameterCodec,
		)
		executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			fmt.Printf("Failed to execute remote coomand\n")
			return err
		}
		return executor.Stream(remotecommand.StreamOptions{
			Stdin:             streamOption.In,
			Stdout:            streamOption.Out,
			Stderr:            streamOption.ErrOut,
			TerminalSizeQueue: t.MonitorSize(t.GetSize()),
		})
	}
	if err := t.Safe(fn); err != nil {
		fmt.Printf("Failed to start tty function\n")
		return err
	}
	return nil
}

func SignalExit(c chan os.Signal, clientset *kubernetes.Clientset, namespace string, podName string) {
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			DeleteDebugPod(clientset, namespace, podName)
		}
	}
}

func newDebugContainer() *v1.Container {
	prop := v1.MountPropagationBidirectional
	priveleged := true
	container := &v1.Container{
		Name:            "debug-bash",
		Image:           "amd64/bash",
		ImagePullPolicy: v1.PullIfNotPresent,
		Command: []string{
			"sleep",
			"86400",
		},
		SecurityContext: &v1.SecurityContext{
			Privileged: &priveleged,
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "docker",
				MountPath: "/var/run/docker.sock",
			},
			{
				Name:      "cgroup",
				MountPath: "/sys/fs/cgroup",
			},
			{
				Name:             "lxcfs",
				MountPath:        "/var/lib/lxc/lxcfs",
				MountPropagation: &prop,
			},
		},
	}
	return container
}
