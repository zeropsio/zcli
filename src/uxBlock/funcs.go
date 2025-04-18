package uxBlock

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/terminal"
)

type GetFunc[T any]func(model tea.Model) (T, error)

func RunR[T any](
	model tea.Model,
	get GetFunc[T],
	opts ...tea.ProgramOption,
) (T, error) {
	if !terminal.IsTerminal() {
		var t T
		return t, errors.Errorf("only allowed in terminal")
	}
	// opts = append(opts, tea.WithAltScreen())
	model, err := tea.NewProgram(model, opts...).Run()
	if err != nil {
		var t T
		return t, err
	}
	return get(model)
}
