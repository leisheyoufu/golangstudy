package k8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/leisheyoufu/golangstudy/k8spodready/pkg/common"
	yaml "gopkg.in/yaml.v2"
)

type containerStatus int

const (
	containerInit  = -1
	containerReady = 1
	containerError = 0
)

var (
	statusMessage = map[containerStatus]string{
		containerInit:  "init",
		containerReady: "ready",
		containerError: "error",
	}
)

func getServices(confFile string) ([]string, error) {
	b, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read conf file %s err=%v\n", confFile, err)
		return nil, err
	}
	svcMap := make(map[string]interface{})
	err = yaml.Unmarshal(b, svcMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal file %s err=%v\n", confFile, err)
		return nil, err
	}
	services, ok := svcMap["services"]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unsupport configuration format, file=%s\n", confFile)
		return nil, err
	}
	ret := common.InterfaceToStringSlice(services)
	return ret, nil
}

func checkServiceReady(kubeconfig string, namespace string, confFile string) bool {
	services, err := getServices(confFile)
	if err != nil {
		return false
	}

	podStatus, err := getPodsStatus(kubeconfig, namespace)
	if err != nil {
		return false
	}
	status := make(map[string]containerStatus)
	for _, service := range services {
		status[service] = containerInit
	}
	for _, service := range services {
		for k, v := range podStatus {
			if strings.HasPrefix(k, service) {
				if status[service] == containerInit && v {
					status[service] = containerReady
				} else if v != true {
					status[service] = containerError
				}
			}
		}
	}

	flag := true
	for k, v := range status {
		if v != containerReady {
			fmt.Fprintf(os.Stderr, "Pod %s: %s, not running.\n", k, statusMessage[v])
			flag = false
		}
	}
	return flag
}
