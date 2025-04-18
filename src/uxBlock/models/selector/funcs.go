package selector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
)

func GetOneSelectedFunc(model tea.Model) (int, error) {
	m, ok := model.(*RootModel)
	if !ok {
		return 0, errors.New("invalid model type") // FIXME lhellmann: type models invalid model error
	}
	if m.Err() != nil {
		return 0, m.Err()
	}
	if m.IsMultiSelect() {
		return 0, errors.New("trying to return only one value") // FIXME lhellmann: error message
	}
	return m.Selected()[0], nil
}

func GetMultipleSelectedFunc(model tea.Model) ([]int, error) {
	m, ok := model.(*RootModel)
	if !ok {
		return nil, errors.New("invalid model type") // FIXME lhellmann: type models invalid model error
	}
	if m.Err() != nil {
		return nil, m.Err()
	}
	if !m.IsMultiSelect() {
		return nil, errors.New("cannot get multiple select from one select") // FIXME lhellmann: error message
	}
	return m.Selected(), nil
}
