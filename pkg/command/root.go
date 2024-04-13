// SPDX-License-Identifier: MIT

package command

import (
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	return &cobra.Command{
		Use:   "kubectl-dpm",
		Short: "kubectl debug profile foo",
		// Long:  "eBPF based road to production identifier traces execve events and help teams to identify the road to production",
	}
}
