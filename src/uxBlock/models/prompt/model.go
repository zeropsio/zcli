package prompt

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type Option = gn.Option[Model]

func WithCursorPosition(cur int) Option {
	return func(m *Model) {
		m.cursor = max(min(cur, len(m.choices)-1), 0)
	}
}

func WithActiveDialogButtonStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.activeDialogButtonStyle = style
	}
}

func WithDialogButtonStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.dialogButtonStyle = style
	}
}

func WithDialogBoxStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.dialogBoxStyle = style
	}
}

func WithQuestionPromptStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.questionPromptStyle = style
	}
}

var defaultQuestionPromptStyle = lipgloss.NewStyle().
	Width(50).
	Align(lipgloss.Center)

type Model struct {
	keyMap  KeyMap
	message string
	choices []string
	cursor  int

	activeDialogButtonStyle lipgloss.Style
	dialogButtonStyle       lipgloss.Style
	dialogBoxStyle          lipgloss.Style
	questionPromptStyle     lipgloss.Style
}

func New(message string, choices []string, opts ...Option) *Model {
	return gn.ApplyOptionsWithDefault(
		Model{
			keyMap:  DefaultKeymap(),
			message: message,
			choices: choices,

			activeDialogButtonStyle: styles.ActiveDialogButton(),
			dialogButtonStyle:       styles.DialogButton(),
			dialogBoxStyle:          styles.DialogBox(),
			questionPromptStyle:     defaultQuestionPromptStyle,
		},
		opts...,
	)
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keyMap.Left):
			m.Left()
		case key.Matches(keyMsg, m.keyMap.Right):
			m.Right()
		}
	}
	return m, nil
}

func (m *Model) Left() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *Model) Right() {
	if m.cursor < len(m.choices)-1 {
		m.cursor++
	}
}

func (m *Model) View() string {
	var buttonsTexts []string
	for i, choice := range m.choices {
		if i == m.cursor {
			buttonsTexts = append(buttonsTexts, m.activeDialogButtonStyle.Render(choice))
		} else {
			buttonsTexts = append(buttonsTexts, m.dialogButtonStyle.Render(choice))
		}
	}

	question := m.questionPromptStyle.Render(m.message)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, buttonsTexts...)
	ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)

	dialog := lipgloss.Place(
		0,
		0,
		lipgloss.Left,
		lipgloss.Center,
		m.dialogBoxStyle.Render(ui),
	)

	return dialog
}
