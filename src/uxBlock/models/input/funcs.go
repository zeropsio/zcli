package input

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/uxBlock/models"
)

func GetValueFunc(model tea.Model) (string, error) {
	input, ok := model.(*RootModel)
	if !ok {
		return "", models.ErrInvalidModelType
	}
	if input.Err() != nil {
		return "", input.Err()
	}
	return input.Value(), nil
}
