package selector

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	LineUp   key.Binding
	LineDown key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	Select      key.Binding
	MultiSelect key.Binding
	SelectAll   key.Binding
	DeselectAll key.Binding

	Filter      key.Binding
	FilterClear key.Binding
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
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "press to jump to the top of the table"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "press to jump to the bottom of the table"),
		),

		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "press to confirm selection"),
		),
		MultiSelect: key.NewBinding(
			key.WithKeys(" "),
			key.WithDisabled(),
			key.WithHelp("spacebar", "press to select multiple"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithDisabled(),
			key.WithHelp("ctrl+a", "press to select all"),
		),
		DeselectAll: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithDisabled(),
			key.WithHelp("ctrl+d", "press to deselect all"),
		),

		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithDisabled(),
			key.WithHelp("/", "press to filter"),
		),
		FilterClear: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithDisabled(),
			key.WithHelp("ctrl+x", "press to clear filter"),
		),
	}
}

type RooKeymap struct {
	Quit key.Binding
	Help key.Binding
}

func DefaultRootKeymap() RooKeymap {
	return RooKeymap{
		Quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc|ctrl+c", "press to quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?|question-mark", "press to show this help"),
		),
	}
}
