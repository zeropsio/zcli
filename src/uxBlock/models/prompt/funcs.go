package prompt

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/uxBlock/models"
)

func GetChoiceCursor(model tea.Model) (int, error) {
	m, ok := model.(*RootModel)
	if !ok {
		return 0, models.ErrInvalidModelType
	}
	if m.Err() != nil {
		return 0, m.Err()
	}
	return m.cursor, nil
}

func GetChoiceValue(model tea.Model) (string, error) {
	m, ok := model.(*RootModel)
	if !ok {
		return "", models.ErrInvalidModelType
	}
	if m.Err() != nil {
		return "", m.Err()
	}
	return m.choices[m.cursor], nil
}
