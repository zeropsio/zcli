package uxBlock

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/terminal"
)

type GetFunc[T any] func(model tea.Model) (T, error)

// RunR runs tea.Model and returns value based on GetFunc[T]
func RunR[T any](
	model tea.Model,
	get GetFunc[T],
	opts ...tea.ProgramOption,
) (T, error) {
	if !terminal.IsTerminal() {
		var t T
		return t, errors.Errorf("allowed only in interactive terminal")
	}
	model, err := tea.NewProgram(model, opts...).Run()
	if err != nil {
		var t T
		return t, err
	}
	return get(model)
}

func getEmpty(tea.Model) (struct{}, error) { return struct{}{}, nil }

// Run runs tea.Model
func Run(
	model tea.Model,
	opts ...tea.ProgramOption,
) error {
	_, err := RunR[struct{}](model, getEmpty, opts...)
	return err
}
