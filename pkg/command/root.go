// SPDX-License-Identifier: MIT

package command

import (
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	return &cobra.Command{
		Use:   "kubectl-dpm",
		Short: "kubectl debug profile manager",
	}
}
