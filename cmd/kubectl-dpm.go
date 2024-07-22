// SPDX-License-Identifier: MIT

package main

import (
	"os"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/bavarianbidi/kubectl-dpm/pkg/command"
	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-dpm", pflag.ExitOnError)
	pflag.CommandLine = flags

	// create root command
	root := command.Root()

	root.PersistentFlags().StringVar(
		&config.ConfigurationFile,
		"config",
		os.Getenv("HOME")+"/.kube-dpm/debug-profiles.yaml",
		"config path",
	)

	// run sub command
	root.AddCommand(
		command.NewCmdDebugProfile(
			genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
		),
	)
	// validation sub command
	root.AddCommand(command.ValidateDebugProfileFile())
	// list sub command
	root.AddCommand(command.List())
	// version sub command
	root.AddCommand(command.Version())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
