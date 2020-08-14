package k8s

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	serviceFile  string
	namespace    string
	checkTimeout int
	Kubeconfig   string
)

func K8sCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "check",
		Long:  `check cluster status`,
		Run:   k8sCheck,
	}
	cmd.Flags().StringVarP(&serviceFile, "service", "s", "services.yml", "servcie file")
	cmd.Flags().IntVarP(&checkTimeout, "timeout", "t", 120, "timeout in second")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "k8s namespace")

	return cmd
}

func k8sCheck(cmd *cobra.Command, args []string) {
	err := wait.Poll(3*time.Second, time.Duration(checkTimeout)*time.Second, func() (bool, error) {
		success := checkServiceReady(Kubeconfig, namespace, serviceFile)
		if !success {
			return false, nil // false and no error will retry
		}
		return true, nil
	})
	if err != nil {
		os.Exit(1)
	}
}
