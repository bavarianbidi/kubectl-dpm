// SPDX-License-Identifier: MIT

package command

import (
	"log"

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

			log.Printf("all profiles are valid\n")
			log.Printf("kubectl path: %s\n", profile.Config.KubectlPath)
			log.Printf("profiles: %v\n", profile.Config.Profiles)

			return nil
		},
	}
}
