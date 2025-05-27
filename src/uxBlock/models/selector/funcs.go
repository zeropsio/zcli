package selector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/uxBlock/models"
)

func GetOneSelectedFunc(model tea.Model) (int, error) {
	m, ok := model.(*RootModel)
	if !ok {
		return 0, models.ErrInvalidModelType
	}
	if m.Err() != nil {
		return 0, m.Err()
	}
	if m.IsMultiSelect() {
		return 0, errors.New("unexpected multiselect in singular value getter")
	}
	return m.Selected()[0], nil
}

func GetMultipleSelectedFunc(model tea.Model) ([]int, error) {
	m, ok := model.(*RootModel)
	if !ok {
		return nil, models.ErrInvalidModelType
	}
	if m.Err() != nil {
		return nil, m.Err()
	}
	return m.Selected(), nil
}
