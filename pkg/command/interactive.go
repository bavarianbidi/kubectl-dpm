// SPDX-License-Identifier: MIT

package command

import (
	bubbletable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
	"github.com/bavarianbidi/kubectl-dpm/pkg/table"
)

type model struct {
	table bubbletable.Model
}

// initTeaModel initializes the model for the interactive mode
func initTeaModel() (model, error) {
	// generate a list of profiles which work in interactive mode
	interactiveProfiles, err := profile.InteractiveProfiles()
	if err != nil {
		return model{}, err
	}

	// generate the table with image, namespace and matchLabels columns
	t := table.GenerateTable(interactiveProfiles, true)

	// read the style config from profile and apply it
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
	t.SetStyles(s)

	return model{
		table: t,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			return m, tea.Quit
		case tea.KeyEnter:
			// set the selected profile name
			flagProfileName = m.table.SelectedRow()[0]

			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.table.View() + "\n  " + m.table.HelpView() + "\n"
}
