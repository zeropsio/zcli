package selector

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	LineUp   key.Binding
	LineDown key.Binding
	PageUp   key.Binding
	PageDown key.Binding

	Select      key.Binding
	MultiSelect key.Binding

	Filter key.Binding
}

func DefaultKeymap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "press to move line up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "press to move line down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("page up", "press to jump up a few lines"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("page down", "press to jump down a few lines"),
		),

		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "press to select one"),
		),
		MultiSelect: key.NewBinding(
			key.WithKeys(" "),
			key.WithDisabled(),
			key.WithHelp("spacebar", "press to select multiple"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithDisabled(),
			key.WithHelp("/", "press to filter"),
		),
	}
}

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
