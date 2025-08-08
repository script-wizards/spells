package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct{}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	return "Spells Tracking TUI\n\n" +
		"Placeholder panes:\n" +
		"- Sessions\n" +
		"- Characters\n" +
		"- Spells\n\n" +
		"Press Ctrl+C to quit"
}