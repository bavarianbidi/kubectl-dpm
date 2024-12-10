// SPDX-License-Identifier: MIT

package table

import (
	"encoding/json"
	"log"

	"github.com/charmbracelet/bubbles/table"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func GenerateTable(profiles []profile.Profile, wide bool) table.Model {
	// generate the empty table
	rows := []table.Row{}

	// find the longest name and profile
	longestName := 0
	longestProfile := 0
	longestImage := 0
	longestNamespace := 0
	longestMatchLabels := 0

	// get all profiles and get string length for column width
	for _, p := range profiles {
		if len(p.ProfileName) > longestName {
			longestName = len(p.ProfileName)
		}

		if len(p.Profile) > longestProfile {
			longestProfile = len(p.Profile)
		}

		if len(p.Image) > longestImage {
			longestImage = len(p.Image)
		}

		if len(p.Namespace) > longestNamespace {
			longestNamespace = len(p.Namespace)
		}

		jsonString, err := json.Marshal(p.MatchLabels)
		if err != nil {
			log.Printf("error marshalling matchLabels: %v", err)
			continue
		}
		if len(string(jsonString)) > longestMatchLabels {
			longestMatchLabels = len(string(jsonString))
		}

		if wide {
			rows = append(rows, table.Row{
				p.ProfileName,
				p.Profile,
				p.Image,
				p.Namespace,
				string(jsonString),
			})
		} else {
			rows = append(rows, table.Row{
				p.ProfileName,
				p.Profile,
			})
		}
	}
	// define the columns
	columns := []table.Column{
		{Title: "Name", Width: longestName},
		{Title: "Profile", Width: longestProfile},
	}

	// add more columns if -w/--wide flag is set
	if wide {
		columns = append(columns, table.Column{Title: "Image", Width: longestImage})
		columns = append(columns, table.Column{Title: "Namespace", Width: longestNamespace})
		columns = append(columns, table.Column{Title: "MatchLabels", Width: longestMatchLabels})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return t
}
