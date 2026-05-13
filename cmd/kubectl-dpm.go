// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/bavarianbidi/kubectl-dpm/pkg/command"
	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	flags := pflag.NewFlagSet("kubectl-dpm", pflag.ExitOnError)
	pflag.CommandLine = flags

	// create root command
	root := command.Root()

	root.PersistentFlags().StringVarP(
		&config.ConfigurationFile,
		"config",
		"c",
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

	return root.ExecuteContext(ctx)
}
