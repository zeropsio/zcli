package logView

import (
	"bytes"
	"strconv"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type Option = generic.Option[Model]

func WithVerticalOffset(offset int) Option {
	return func(c *Model) {
		c.VerticalOffset = offset
	}
}

func WithEnabled(e bool) Option {
	return func(c *Model) {
		c.Enabled = e
	}
}

func WithFollow(f bool) Option {
	return func(c *Model) {
		c.Follow = f
	}
}

type Model struct {
	buffer     *bytes.Buffer
	viewport   viewport.Model
	spinner    spinner.Model
	lastBufLen int

	Enabled        bool
	VerticalOffset int
	Follow         bool

	mu      sync.Mutex
	cmdSink *models.CmdSink
}

func New(width, height int, options ...Option) *Model {
	return generic.ApplyOptionsWithDefault(
		Model{
			buffer:   new(bytes.Buffer),
			viewport: viewport.New(width, height),
			spinner:  spinner.New(spinner.WithSpinner(spinner.MiniDot)),
			cmdSink:  models.NewCmdSink(),
			Follow:   true,
		},
		options...,
	)
}

func (m *Model) Write(p []byte) (n int, err error) {
	return m.buffer.Write(p)
}

const spacebar = " "

func (m *Model) Init() tea.Cmd {
	// default key map uses 'f' key for pgDown, we need it to enable/disable follow mode
	m.viewport.KeyMap.PageDown = key.NewBinding(
		key.WithKeys("pgdown", spacebar),
		key.WithHelp("pgdn", "page down"),
	)

	if m.Enabled {
		return m.spinner.Tick
	}
	return nil
}

func (m *Model) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Enabled = true
	m.cmdSink.Pour(m.spinner.Tick)
}

func (m *Model) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Enabled = false
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cmdSink.Filled() {
		return m, m.cmdSink.DrainBatch()
	}

	if resize, isResize := msg.(tea.WindowSizeMsg); isResize {
		headerHeight := lipgloss.Height(m.followText())
		m.viewport.Width = resize.Width
		m.viewport.Height = resize.Height - m.VerticalOffset - headerHeight
		m.viewport.YOffset = headerHeight
	}
	if !m.Enabled {
		return m, m.cmdSink.DrainBatch()
	}

	m.viewport.SetContent(m.buffer.String())
	if keyMsg, isKeyMsg := msg.(tea.KeyMsg); isKeyMsg {
		if keyMsg.String() == "f" {
			m.Follow = !m.Follow
		}
	}

	// follow logic
	if m.Follow && m.buffer.Len() != m.lastBufLen {
		m.viewport.GotoBottom()
		m.lastBufLen = m.buffer.Len()
	}

	m.spinner = models.Update[spinner.Model](m.cmdSink, msg, m.spinner)
	m.viewport = models.Update[viewport.Model](m.cmdSink, msg, m.viewport)

	return m, m.cmdSink.DrainBatch()
}

func (m *Model) View() string {
	if !m.Enabled {
		return ""
	}
	s := m.followText()
	if m.buffer.Len() == 0 {
		s += m.spinner.View() + " " + styles.InfoText("Waiting for logs").String()
	} else {
		s += "\n" + m.viewport.View()
	}
	return s
}

// followText returns colored info text
// 'follow: %t (press 'f' to start/stop following)\n'
func (m *Model) followText() string {
	b := styles.NewStringBuilder()
	b.WriteInfoColor("follow: ")
	b.WriteSelectColor(strconv.FormatBool(m.Follow))
	b.WriteInfoColor(" (press '")
	b.WriteSelectColor("f")
	b.WriteInfoColor("' to start/stop following)")
	b.WriteRune('\n')
	return b.String()
}
