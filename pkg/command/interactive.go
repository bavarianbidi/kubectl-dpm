package command

import (
	"fmt"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type model struct {
	profiles []profile.Profile
	cursor   int // which to-do list item our cursor is pointing at
	table    table.Model
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func initialModel() model {

	if term.IsTerminal(0) {
		println("in a term")
	} else {
		println("not in a term")
	}
	width, height, _ := term.GetSize(0)
	fmt.Printf("width: %d, height: %d\n", width, height)

	rows := []table.Row{}

	interactiveProfiles := profile.InteractiveProfiles()

	// find the longest name and profile
	longestName := 0
	longestProfile := 0

	for _, p := range interactiveProfiles {

		if len(p.ProfileName) > longestName {
			longestName = len(p.ProfileName)
		}

		if len(p.Profile) > longestProfile {
			longestProfile = len(p.Profile)
		}

		rows = append(rows, table.Row{
			p.ProfileName,
			p.Profile,
		})
	}
	columns := []table.Column{
		{Title: "Name", Width: longestName},
		{Title: "Profile", Width: longestProfile},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		// table.WithHeight(height-10),
		// table.WithWidth(width-40),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// we only use profiles where everything is defined

	return model{
		// Our to-do list is a grocery list
		profiles: interactiveProfiles,
		table:    t,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c", "ctrl+d":
			return m, tea.Quit
		case "enter":

			flagProfileName = m.table.SelectedRow()[0]

			// run the selected profile
			run([]string{})

			// check the different behavior of the blur()
			m.table.Blur()

			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}
