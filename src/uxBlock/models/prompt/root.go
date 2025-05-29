package prompt

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/uxBlock/models"
)

type RootModel struct {
	//nolint:containedctx
	ctx    context.Context
	cancel context.CancelFunc
	*Model

	quit   bool
	err    error
	keyMap RooKeymap
}

func NewRoot(ctx context.Context, message string, choices []string, opts ...Option) *RootModel {
	ctx, cancel := context.WithCancel(ctx)
	return &RootModel{
		ctx:    ctx,
		cancel: cancel,
		Model:  New(message, choices, opts...),
		keyMap: DefaultRootKeymap(),
	}
}

func (r *RootModel) Err() error {
	return r.err
}

func (r *RootModel) Init() tea.Cmd {
	return nil
}

func (r *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
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
			r.err = models.ErrCtrlC
			return r, models.Noop
		case key.Matches(keyMsg, r.keyMap.Select):
			r.quit = true
			return r, models.Noop
		}
	}

	r.Model, cmd = r.Model.Update(msg)
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

func (r *RootModel) View() string {
	if r.quit {
		return ""
	}
	return r.Model.View()
}
