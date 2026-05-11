// SPDX-License-Identifier: MIT

package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
	table "github.com/bavarianbidi/kubectl-dpm/pkg/table"
)

func List() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all profiles",

		RunE: func(_ *cobra.Command, _ []string) error {
			if err := config.GenerateConfig(); err != nil {
				return fmt.Errorf("generate config: %w", err)
			}

			if err := generateListOutput(); err != nil {
				return fmt.Errorf("generate list output: %w", err)
			}

			return nil
		},
	}

	listCmd.Flags().BoolVarP(&flagVerboseList, "wide", "w", false, "show more information about profiles")

	return listCmd
}

func generateListOutput() error {
	tbl := table.GenerateTable(profile.Config.Profiles, flagVerboseList)
	table.ConfigureStatic(&tbl)

	if _, err := fmt.Fprintln(os.Stdout, tbl.View()); err != nil {
		return fmt.Errorf("print list table: %w", err)
	}

	return nil
}
