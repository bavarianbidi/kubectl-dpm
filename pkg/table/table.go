// SPDX-License-Identifier: MIT

package table

import (
	"encoding/json"
	"log"

	bubbletable "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func GenerateTable(profiles []profile.Profile, wide bool) bubbletable.Model {
	// generate the empty table
	rows := []bubbletable.Row{}

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
			rows = append(rows, bubbletable.Row{
				p.ProfileName,
				p.Profile,
				p.Image,
				p.Namespace,
				string(jsonString),
			})
		} else {
			rows = append(rows, bubbletable.Row{
				p.ProfileName,
				p.Profile,
			})
		}
	}
	// define the columns
	columns := []bubbletable.Column{
		{Title: "Name", Width: longestName},
		{Title: "Profile", Width: longestProfile},
	}

	// add more columns if -w/--wide flag is set
	if wide {
		columns = append(columns, bubbletable.Column{Title: "Image", Width: longestImage})
		columns = append(columns, bubbletable.Column{Title: "Namespace", Width: longestNamespace})
		columns = append(columns, bubbletable.Column{Title: "MatchLabels", Width: longestMatchLabels})
	}

	t := bubbletable.New(
		bubbletable.WithColumns(columns),
		bubbletable.WithRows(rows),
		bubbletable.WithFocused(true),
	)

	return t
}

// ConfigureInteractive applies the focused styling used by the interactive run flow.
func ConfigureInteractive(t *bubbletable.Model) {
	t.Focus()
	t.SetStyles(styles())
}

// ConfigureStatic adapts the shared table model for plain list output.
// It removes cell padding, disables the selected-row highlight, and expands the
// viewport so View prints only the actual table rows without interactive filler.
func ConfigureStatic(t *bubbletable.Model) {
	s := styles()
	s.Header = s.Header.UnsetPadding()
	s.Cell = s.Cell.UnsetPadding()
	s.Selected = s.Cell
	t.SetColumns(staticColumns(t.Columns()))
	t.SetStyles(s)
	t.Blur()
	t.SetHeight(len(t.Rows()) + lipgloss.Height(t.View()) - t.Height())
}

// styles builds the common Bubble Tea table styles from the configured theme.
func styles() bubbletable.Styles {
	profile.CompleteStyle()

	s := bubbletable.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		Background(lipgloss.Color(profile.Config.Style.HeaderBackgroundColor)).
		Foreground(lipgloss.Color(profile.Config.Style.HeaderForegroundColor)).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(profile.Config.Style.SelectedForegroundColor)).
		Background(lipgloss.Color(profile.Config.Style.SelectedBackgroundColor)).
		Bold(false)

	return s
}

// staticColumns widens every non-terminal column by one cell to preserve a
// visible gap after removing Bubble Tea's default left/right cell padding.
func staticColumns(columns []bubbletable.Column) []bubbletable.Column {
	staticColumns := make([]bubbletable.Column, len(columns))
	copy(staticColumns, columns)

	for i := 0; i < len(staticColumns)-1; i++ {
		staticColumns[i].Width++
	}

	return staticColumns
}
