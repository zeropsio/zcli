package input

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
)

func GetValueFunc(model tea.Model) (string, error) {
	input, ok := model.(*RootModel)
	if !ok {
		return "", errors.New("invalid model type") // FIXME lhellmann: type models invalid model error
	}
	if input.Err() != nil {
		return "", input.Err()
	}
	return input.Value(), nil
}
