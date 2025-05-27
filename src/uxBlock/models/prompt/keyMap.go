package prompt

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Left  key.Binding
	Right key.Binding
}

func DefaultKeymap() KeyMap {
	return KeyMap{
		Left: key.NewBinding(
			key.WithKeys("left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
		),
	}
}

type RooKeymap struct {
	Quit   key.Binding
	Select key.Binding
}

func DefaultRootKeymap() RooKeymap {
	return RooKeymap{
		Quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
		),
	}
}
