// SPDX-License-Identifier: MIT

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func ValidateDebugProfileFile() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "validate debug profiles configuration file",

		RunE: func(c *cobra.Command, args []string) error {
			if err := config.GenerateConfig(); err != nil {
				return err
			}

			if err := profile.ValidateAllProfiles(); err != nil {
				return err
			}

			fmt.Printf("all profiles are valid\n")

			return nil
		},
	}
}
