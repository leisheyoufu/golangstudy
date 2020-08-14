package main

import (
	"fmt"
	"os/user"
	"path"

	"os"

	"github.com/leisheyoufu/golangstudy/k8spodready/k8s"
	"github.com/spf13/cobra"
)

const (
	cliName        = "k8sready"
	cliDescription = "A simple command line tool to test if pods are ready"
)

var (
	Version    string
	BuildTime  string
	Commit     string
	kubeconfig string
	config     string
)

func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of k8sready",
		Long:  `Print the version number of k8sready`,
		Run:   version,
	}
	return cmd
}

func version(cmd *cobra.Command, args []string) {
	fmt.Printf("Version: %s, BuildTime: %s\n Commit: %s\n", Version, BuildTime, Commit)
}

func initConfig() {
	k8s.Kubeconfig = kubeconfig
}

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	defaultConfPath := path.Join(user.HomeDir, ".kube", "config")
	cmd := &cobra.Command{
		Use:   cliName,
		Short: cliDescription,
		Long: `k8sready --help and k8sready help COMMAND to see the usage for specfied
	command.`,
		SuggestFor: []string{"k8sready"},
	}
	cmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "", defaultConfPath, "k8s kubeconfig file")
	cobra.OnInitialize(initConfig)
	cmd.AddCommand(versionCommand())
	cmd.AddCommand(k8s.K8sCheckCommand())

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
