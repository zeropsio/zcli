package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zcli/src/optional"
)

type Option = generic.Option[Model]

func WithLabel(label string) Option {
	return func(m *Model) {
		m.label = optional.New(label)
	}
}

func WithLabelStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.labelStyle = style
	}
}

func WithSetSuggestions(suggestions []string) Option {
	return func(m *Model) {
		m.input.ShowSuggestions = true
		m.input.SetSuggestions(suggestions)
	}
}

func WithPlaceholder(placeholder string) Option {
	return func(m *Model) {
		m.input.Placeholder = placeholder
	}
}

func WithHelpPlaceholder() Option {
	return func(m *Model) {
		m.input.Placeholder = "press '?' to show help"
	}
}

func WithPlaceholderStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.input.PlaceholderStyle = style
	}
}

func WithoutPrompt() Option {
	return func(m *Model) {
		m.input.Prompt = ""
	}
}

func WithPromptStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.input.PromptStyle = style
	}
}

func WithPrompt(prompt string) Option {
	return func(m *Model) {
		m.input.Prompt = prompt
	}
}

func WithMinMaxWidth(min, max int) Option {
	return func(m *Model) {
		m.minWidth, m.maxWidth = min, max
	}
}

type Model struct {
	label      optional.Null[string]
	labelStyle lipgloss.Style
	input      textinput.Model

	minWidth, maxWidth int
}

func New(opts ...Option) *Model {
	return generic.ApplyOptionsWithDefault(
		Model{
			input: textinput.New(),
		},
		opts...,
	)
}

func (m *Model) Value() string {
	return m.input.Value()
}

func (m *Model) Focus() tea.Cmd {
	return m.input.Focus()
}

func (m *Model) Init() tea.Cmd {
	m.input.KeyMap.AcceptSuggestion.SetHelp("tab", "press to accept suggestion")
	m.input.KeyMap.NextSuggestion.SetHelp("up", "press to select next suggestion")
	m.input.KeyMap.PrevSuggestion.SetHelp("down", "press to select previous suggestion")
	m.input.KeyMap.AcceptSuggestion.SetEnabled(m.input.ShowSuggestions)
	m.input.KeyMap.NextSuggestion.SetEnabled(m.input.ShowSuggestions)
	m.input.KeyMap.PrevSuggestion.SetEnabled(m.input.ShowSuggestions)
	return nil
}

func (m *Model) Resize(width int) {
	labelWidth := lipgloss.Width(m.label.OrDefault(""))
	width = width - labelWidth - 1
	width = max(width, m.minWidth)
	width = min(width, m.maxWidth)
	m.input.Width = width
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	var label string
	if m.label.Filled() {
		label = m.labelStyle.Render(m.label.Some()) + ": "
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, label, m.input.View())
}
