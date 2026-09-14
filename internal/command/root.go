// SPDX-License-Identifier: MIT

package command

import (
	"github.com/spf13/cobra"
)

// Root creates and returns the root cobra command.
func Root() *cobra.Command {
	return &cobra.Command{
		Use:           "kubectl-dpm",
		Short:         "kubectl debug profile manager",
		SilenceUsage:  false,
		SilenceErrors: false,
	}
}
