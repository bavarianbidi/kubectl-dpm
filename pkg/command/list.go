// SPDX-License-Identifier: MIT

package command

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"

	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func List() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all profiles",

		RunE: func(c *cobra.Command, args []string) error {
			if err := config.GenerateConfig(); err != nil {
				return err
			}

			if err := generateListOutput(); err != nil {
				return err
			}

			return nil
		},
	}

	listCmd.Flags().BoolVarP(&flagVerboseList, "wide", "w", false, "show more information about profiles")

	return listCmd
}

func generateListOutput() error {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	var tbl table.Table
	tbl = table.New("Name", "Profile")

	if flagVerboseList {
		tbl = table.New("Name", "Profile", "Image", "Namespace", "MatchLabels")
	}
	tbl.
		WithHeaderFormatter(headerFmt).
		WithFirstColumnFormatter(columnFmt)

	for _, p := range profile.Config.Profiles {
		var matchLabels string
		for label, value := range p.MatchLabels {
			matchLabels += fmt.Sprintf("%s=%s, ", label, value)
		}
		tbl.AddRow(
			p.ProfileName,       // Name
			p.CustomProfileFile, // Profile
			p.Image,             // Image
			p.Namespace,         // Namespace
			matchLabels,         // MatchLabels
		)
	}

	tbl.Print()

	return nil
}
