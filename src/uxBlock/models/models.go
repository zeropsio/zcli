package models

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// Noop forces bubble tea to render
func Noop() tea.Msg {
	return struct{}{}
}

var ErrCtrlC = errors.New("ctrl+c")

type Helper interface {
	Enabled() bool
	Help() key.Help
}

func FormatHelp(helpers ...Helper) string {
	t := table.New().
		Border(lipgloss.Border{}).
		BorderTop(false).
		BorderHeader(false)
	rows := make([][]string, 0, len(helpers))
	for _, helper := range helpers {
		if !helper.Enabled() {
			continue
		}
		h := helper.Help()
		rows = append(rows, []string{h.Key, h.Desc})
	}
	t.Rows(rows...)
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if col == 1 {
			return styles.HelpStyle().Padding(0, 1)
		}
		return styles.HelpStyle()
	})
	return t.Render()
}
