// SPDX-License-Identifier: MIT

package command

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
	table "github.com/bavarianbidi/kubectl-dpm/pkg/table"
)

func List() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all profiles",

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.GenerateConfig(); err != nil {
				return fmt.Errorf("generate config: %w", err)
			}

			if err := generateListOutput(cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("generate list output: %w", err)
			}

			return nil
		},
	}

	listCmd.Flags().BoolVarP(&flagVerboseList, "wide", "w", false, "show more information about profiles")

	return listCmd
}

func generateListOutput(w io.Writer) error {
	tbl := table.GenerateTable(profile.Config.Profiles, flagVerboseList)
	table.ConfigureStatic(&tbl)

	if _, err := fmt.Fprintln(w, tbl.View()); err != nil {
		return fmt.Errorf("print list table: %w", err)
	}

	return nil
}
