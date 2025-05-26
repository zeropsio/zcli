package selector

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type RootModel struct {
	//nolint:containedctx
	ctx    context.Context
	cancel context.CancelFunc
	*Model

	showHelp      bool
	quit          bool
	err           error
	keyMap        RooKeymap
	width, height int
}

func NewRoot(ctx context.Context, tableBody *table.Body, opts ...Option) *RootModel {
	ctx, cancel := context.WithCancel(ctx)
	return &RootModel{
		ctx:    ctx,
		cancel: cancel,
		Model:  New(tableBody, opts...),
		keyMap: DefaultRootKeymap(),
	}
}

func (r *RootModel) Err() error {
	return r.err
}

func (r *RootModel) Init() tea.Cmd {
	return r.Model.Init()
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

	if resizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		heightOffset := 0
		if r.showHelp {
			heightOffset = lipgloss.Height(r.HelpView())
		}
		r.Model.Resize(
			resizeMsg.Width,
			resizeMsg.Height-heightOffset,
		)
		r.width = resizeMsg.Width
		r.height = resizeMsg.Height
		cmds = append(cmds, models.Noop)
	}

	if _, ok := msg.(SelectedMsg); ok {
		r.quit = true
		return r, models.Noop
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, r.keyMap.Quit):
			r.quit = true
			r.err = models.ErrCtrlC
			return r, models.Noop
		case key.Matches(keyMsg, r.keyMap.Help):
			r.showHelp = !r.showHelp
			heightOffset := 0
			if r.showHelp {
				heightOffset = lipgloss.Height(r.HelpView())
			}
			r.Model.Resize(r.width, r.height-heightOffset)
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
	if r.showHelp {
		return lipgloss.JoinVertical(lipgloss.Left,
			r.Model.View(),
			styles.HelpStyle().Render("press '?' to hide help"),
			r.HelpView(),
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		r.Model.View(),
		styles.HelpStyle().Render("press '?' to show help"),
	)
}

func (r *RootModel) HelpView() string {
	return models.FormatHelp(
		r.Model.keyMap.LineUp,
		r.Model.keyMap.LineDown,
		r.Model.keyMap.PageUp,
		r.Model.keyMap.PageDown,
		r.Model.keyMap.MultiSelect,
		r.Model.keyMap.SelectAll,
		r.Model.keyMap.DeselectAll,
		r.Model.keyMap.Filter,
		r.Model.keyMap.FilterClear,
		r.Model.keyMap.Select,
		r.keyMap.Quit,
	)
}
