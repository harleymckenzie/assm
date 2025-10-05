package table

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table      table.Model
	selectedID string
	done       bool
}

func (m model) Init() tea.Cmd { return nil }

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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if len(m.table.SelectedRow()) > 1 {
				m.selectedID = m.table.SelectedRow()[1]
				m.done = true
				return m, tea.Quit
			}
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.done {
		return ""
	}
	return baseStyle.Render(m.table.View()) + "\n"
}

func ShowTableAndSelect(rows []table.Row) string {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Instance ID", Width: 20},
		{Title: "State", Width: 10},
		{Title: "Type", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("33")).
		Bold(false)
	t.SetStyles(s)

	m := model{table: t}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	// Cast the final model to get the selectedID
	if finalM, ok := finalModel.(model); ok {
		return finalM.selectedID
	}
	
	return ""
}