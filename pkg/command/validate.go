// SPDX-License-Identifier: MIT

package command

import (
	"github.com/spf13/cobra"

	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func ValidateDebugProfileFile() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "validate debug profiles configuration file",

		RunE: func(_ *cobra.Command, _ []string) error {
			if err := config.GenerateConfig(); err != nil {
				return err
			}

			if err := profile.ValidateAllProfiles(); err != nil {
				return err
			}

			if err := generateListOutput(); err != nil {
				return err
			}

			return nil
		},
	}
}
