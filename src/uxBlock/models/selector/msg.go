package selector

import (
	tea "github.com/charmbracelet/bubbletea"
)

type SelectedMsg struct{}

func Selected() tea.Msg {
	return SelectedMsg{}
}
