package deviceLogin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/dto/output"
)

func runInteractive(ctx context.Context, f flow, cfg Config) (output.AuthLogin, error) {
	m := newModel(ctx, f, cfg)
	defer m.cancel()
	final, err := tea.NewProgram(m, tea.WithoutSignalHandler(), tea.WithContext(ctx)).Run()
	if err != nil {
		if ctx.Err() != nil {
			return output.AuthLogin{}, ctx.Err()
		}
		return output.AuthLogin{}, err
	}
	return result(final)
}

func result(final tea.Model) (output.AuthLogin, error) {
	m, ok := final.(*model)
	if !ok {
		return output.AuthLogin{}, models.ErrInvalidModelType
	}
	if m.err != nil {
		return output.AuthLogin{}, m.err
	}
	return m.auth, nil
}

type state int

const (
	stateInitiating state = iota
	stateWaiting
	stateExpired
	stateDone
)

type keyMap struct {
	quit    key.Binding
	newCode key.Binding
}

// The code box and the URL go to scrollback with tea.Println. The renderer truncates live lines to the terminal width, which would cut the URL.
// The live area holds only the spinner line or the expired prompt.
type model struct {
	//nolint:containedctx
	ctx    context.Context
	cancel context.CancelFunc
	flow   flow
	cfg    Config
	keyMap keyMap

	spinner spinner.Model
	state   state
	// generation tags the messages of one request, so messages of a request replaced after an expiry are dropped.
	generation int
	req        request
	deadline   time.Time

	auth output.AuthLogin
	err  error
	quit bool
}

type initiatedMsg struct {
	generation int
	req        request
	err        error
}

type browserMsg struct {
	generation int
	line       string
}

type pollTickMsg struct {
	generation int
}

type pollMsg struct {
	generation int
	result     pollResult
}

func newModel(ctx context.Context, f flow, cfg Config) *model {
	ctx, cancel := context.WithCancel(ctx)
	return &model{
		ctx:    ctx,
		cancel: cancel,
		flow:   f,
		cfg:    cfg,
		keyMap: keyMap{
			quit:    key.NewBinding(key.WithKeys("ctrl+c")),
			newCode: key.NewBinding(key.WithKeys("enter")),
		},
		spinner: spinner.New(spinner.WithSpinner(spinner.MiniDot)),
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.initiateCmd())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quit {
		m.cancel()
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.updateKey(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case initiatedMsg:
		if msg.generation != m.generation {
			return m, nil
		}
		return m.updateInitiated(msg)
	case browserMsg:
		if msg.generation != m.generation {
			return m, nil
		}
		return m, tea.Println(msg.line)
	case pollTickMsg:
		if msg.generation != m.generation || m.state != stateWaiting {
			return m, nil
		}
		return m, m.checkCmd()
	case pollMsg:
		if msg.generation != m.generation || m.state != stateWaiting {
			return m, nil
		}
		return m.updatePoll(msg.result)
	}
	return m, nil
}

func (m *model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.quit):
		return m.finish(models.ErrCtrlC)
	case key.Matches(msg, m.keyMap.newCode) && m.state == stateExpired:
		m.generation++
		m.state = stateInitiating
		return m, tea.Batch(tea.Println(i18n.T(i18n.DeviceLoginPreviousExpired)), m.initiateCmd())
	}
	return m, nil
}

func (m *model) updateInitiated(msg initiatedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m.finish(msg.err)
	}
	m.req = msg.req
	m.deadline = time.Now().Add(msg.req.expiresIn)
	m.state = stateWaiting
	return m, tea.Batch(tea.Println(requestText(msg.req)), m.openBrowserCmd(), m.pollTickCmd())
}

func (m *model) updatePoll(res pollResult) (tea.Model, tea.Cmd) {
	switch res.outcome {
	case pollPending:
		return m, m.pollTickCmd()
	case pollRetry:
		m.cfg.UxBlocks.LogDebug("device login: check failed, retrying: " + res.err.Error())
		return m, m.pollTickCmd()
	case pollApproved:
		m.auth = res.auth
		return m.finish(nil)
	case pollDenied:
		return m.finish(deniedError(m.flow))
	case pollExpired:
		m.state = stateExpired
		return m, nil
	case pollFailed:
		return m.finish(res.err)
	}
	return m.finish(errors.Errorf("unexpected poll outcome %d", res.outcome))
}

// finish records the outcome and renders one more frame before the next Update quits, as the other root models do.
func (m *model) finish(err error) (tea.Model, tea.Cmd) {
	m.err = err
	m.state = stateDone
	m.quit = true
	return m, models.Noop
}

func (m *model) initiateCmd() tea.Cmd {
	ctx, f, cfg, generation := m.ctx, m.flow, m.cfg, m.generation
	return func() tea.Msg {
		req, err := initiate(ctx, f, cfg)
		return initiatedMsg{generation: generation, req: req, err: err}
	}
}

func (m *model) openBrowserCmd() tea.Cmd {
	cfg, url, generation := m.cfg, m.req.url, m.generation
	return func() tea.Msg {
		return browserMsg{generation: generation, line: openBrowser(cfg, url)}
	}
}

func (m *model) pollTickCmd() tea.Cmd {
	generation := m.generation
	return tea.Tick(m.req.interval, func(time.Time) tea.Msg {
		return pollTickMsg{generation: generation}
	})
}

func (m *model) checkCmd() tea.Cmd {
	ctx, f, req, generation := m.ctx, m.flow, m.req, m.generation
	return func() tea.Msg {
		return pollMsg{generation: generation, result: check(ctx, f, req)}
	}
}

func (m *model) View() string {
	switch m.state {
	case stateWaiting:
		return m.spinner.View() + " " + i18n.T(i18n.DeviceLoginWaiting, m.countdown()) + "\n"
	case stateExpired:
		return i18n.T(i18n.DeviceLoginExpired) + "\n"
	case stateInitiating, stateDone:
		return ""
	}
	return ""
}

func (m *model) countdown() string {
	seconds := max(int(time.Until(m.deadline)/m.cfg.unit), 0)
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

func requestText(req request) string {
	return i18n.T(i18n.DeviceLoginConfirmCode) + "\n" + codeBox(req.userCode) + "\n" + req.url
}

func codeBox(code string) string {
	return styles.DefaultStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.InfoColor).
		Padding(0, 2).
		Bold(true).
		Foreground(styles.SelectColor).
		Render(spaced(code))
}

func spaced(code string) string {
	letters := strings.Split(code, "")
	return strings.Join(letters, " ")
}
