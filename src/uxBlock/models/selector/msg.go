package selector

import (
	tea "github.com/charmbracelet/bubbletea"
)

type SelectedMsg struct{}

func (m *Model) SelectCommand() tea.Msg {
	if !m.hasResults() {
		return nil
	}
	return SelectedMsg{}
}
