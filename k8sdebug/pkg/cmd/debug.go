package cmd

import (
	"fmt"
	"os"

	"github.com/leisheyoufu/golangstudy/k8sdebug/pkg/client"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/cmd/exec"
)

const (
	defaultImage  = "busybox:1.28.4"
	defaultConfig = "config"
)

var (
	debugExample = `
	# debug container
	%[1]s debug <pod>
	# debug pod container in namespace foo
	%[1]s debug <pod> --namespace foo
	# debug container c1 of pod p1 in namespace pod
	%[1]s debug <pod> --container c1 --namespace foo`
)

type DebugOptions struct {
	exec.StreamOptions
	kubeConfig  string
	Image       string
	Pod         string
	Namespace   string
	rawConfig   api.Config
	RestConfig  *restclient.Config
	args        []string
	KubeCli     *kubernetes.Clientset
	configFlags *genericclioptions.ConfigFlags
}

func NewDebugOptions(streams genericclioptions.IOStreams) *DebugOptions {
	return &DebugOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		StreamOptions: exec.StreamOptions{
			IOStreams: genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
			TTY:       true,
			Stdin:     true,
		},
	}
}

func NewCmdDebug(streams genericclioptions.IOStreams) *cobra.Command {
	opts := NewDebugOptions(streams)
	cmd := &cobra.Command{
		Use:          "debug [pod] [flags]",
		Short:        "Debug pod container",
		Example:      fmt.Sprintf(debugExample, "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := opts.Complete(c, args); err != nil {
				return err
			}
			if err := opts.Validate(); err != nil {
				return err
			}
			if err := opts.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Image, "image", "",
		fmt.Sprintf("Container Image to run the debug container, default to %s", defaultImage))
	cmd.Flags().StringVar(&opts.kubeConfig, "config", "config",
		fmt.Sprintf("Kubernetes config file, default to %s", "config"))
	cmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "default",
		fmt.Sprintf("Kubernetes namespace, default to %s", "default"))

	return cmd
}

// Complete sets all information required for updating the current context
func (o *DebugOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var err error
	o.rawConfig, err = o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *DebugOptions) Validate() error {
	if len(o.args) != 1 {
		return fmt.Errorf("Only one argument is allowed.")
	}
	o.Pod = o.args[0]
	return nil
}

// Check and run debug container inside pod
func (o *DebugOptions) Run() error {
	var err error
	o.KubeCli, o.RestConfig, err = client.GetClientset(o.kubeConfig)
	if err != nil {
		return err
	}
	return client.Debug(o.KubeCli, o.RestConfig, o.StreamOptions, o.Namespace, o.Pod)
}
