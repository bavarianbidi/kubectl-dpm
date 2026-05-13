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

		RunE: func(c *cobra.Command, _ []string) error {
			if err := config.GenerateConfig(); err != nil {
				return fmt.Errorf("generate config: %w", err)
			}

			if err := profile.ValidateAllProfiles(c.Context()); err != nil {
				return fmt.Errorf("validate profiles: %w", err)
			}

			return nil
		},
	}
}
