package main

import (
	"fmt"
	"github.com/leisheyoufu/golangstudy/cos/coscli/coscli"
	"os"

	"github.com/spf13/cobra"
)

const (
	cliName        = "cosctl"
	cliDescription = "A simple command line to test cos"
)

func main() {
	cmd := &cobra.Command{
		Use:   cliName,
		Short: cliDescription,
		Long: `cosctl --help and cosctl help COMMAND to see the usage for specfied
	command.`,
		SuggestFor: []string{"cosctl"},
	}
	cmd.PersistentFlags().StringVarP(&coscli.SecretID, "secret-id", "i", "xxxxx", "secret-id")
	cmd.PersistentFlags().StringVarP(&coscli.SecretKey, "secret-key", "k", "xxxx", "secret-key")
	cmd.AddCommand(coscli.ListBucketCommand())
	cmd.AddCommand(coscli.ListObjectCommand())
	cmd.AddCommand(coscli.GetObjectCommand())

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
