package input

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/uxBlock/models"
)

type RootModel struct {
	ctx    context.Context
	cancel context.CancelFunc
	*Model

	showHelp bool
	quit     bool
	err      error
	keyMap   RooKeymap
}

func NewRoot(ctx context.Context, opts ...Option) *RootModel {
	ctx, cancel := context.WithCancel(ctx)
	return &RootModel{
		ctx:    ctx,
		cancel: cancel,
		Model:  New(opts...),
		keyMap: DefaultRootKeymap(),
	}
}

func (r *RootModel) Err() error {
	return r.err
}

func (r *RootModel) Init() tea.Cmd {
	return tea.Batch(r.Model.Focus(), r.Model.Init())
}

func (r *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if r.quit {
		if r.err != nil {
			r.cancel()
		}
		return r, tea.Quit
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, r.keyMap.Quit):
			r.quit = true
			r.err = models.CtrlC
			return r, models.Noop
		case key.Matches(keyMsg, r.keyMap.Submit):
			r.quit = true
			return r, models.Noop
		case key.Matches(keyMsg, r.keyMap.Help):
			r.showHelp = !r.showHelp
			return r, models.Noop
		}
	}

	if resizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		r.Model.Resize(resizeMsg.Width)
	}

	var cmd tea.Cmd
	r.Model, cmd = r.Model.Update(msg)
	return r, cmd
}

func (r *RootModel) View() string {
	if r.quit {
		return ""
	}
	if r.showHelp {
		return lipgloss.JoinVertical(lipgloss.Left,
			r.Model.View(),
			r.HelpView(),
		)
	}
	return r.Model.View()
}

func (r *RootModel) HelpView() string {
	return models.FormatHelp(
		r.input.KeyMap.AcceptSuggestion,
		r.input.KeyMap.NextSuggestion,
		r.input.KeyMap.PrevSuggestion,
		r.keyMap.Submit,
		r.keyMap.Quit,
	)
}
