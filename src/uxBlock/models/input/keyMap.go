package input

import (
	"github.com/charmbracelet/bubbles/key"
)

type RooKeymap struct {
	Quit   key.Binding
	Submit key.Binding
	Help   key.Binding
}

func DefaultRootKeymap() RooKeymap {
	return RooKeymap{
		Quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc|ctrl+c", "press to quit"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "press to submit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?|question-mark", "press to show this help"),
		),
	}
}
