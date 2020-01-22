package main

import (
	"os"

	"github.com/leisheyoufu/golangstudy/k8sdebug/pkg/cmd"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-debug", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewCmdDebug(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
