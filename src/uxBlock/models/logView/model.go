package logView

import (
	"bytes"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type Option = gn.Option[Model]

func WithVerticalOffset(offset int) Option {
	return func(c *Model) {
		c.verticalOffset = offset
	}
}

func WithMaxHeight(height int) Option {
	return func(c *Model) {
		c.maxHeight = height
		c.Resize()
	}
}

func WithEnabled(e bool) Option {
	return func(c *Model) {
		c.enabled = e
	}
}

func WithFollow(f bool) Option {
	return func(c *Model) {
		c.follow = f
	}
}

func WithAdditionalText(text string) Option {
	return func(c *Model) {
		c.additionalText = text
	}
}

type Model struct {
	buffer     *bytes.Buffer
	viewport   viewport.Model
	spinner    spinner.Model
	lastBufLen int

	enabled        bool
	verticalOffset int
	follow         bool
	maxHeight      int
	width, height  int

	additionalText string

	cmds []tea.Cmd
}

func New(options ...Option) *Model {
	return gn.ApplyOptionsWithDefault(
		Model{
			buffer:   new(bytes.Buffer),
			viewport: viewport.New(0, 0),
			spinner:  spinner.New(spinner.WithSpinner(spinner.MiniDot)),
			follow:   true,
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

	if m.enabled {
		return m.spinner.Tick
	}
	return nil
}

func (m *Model) Enabled() bool {
	return m.enabled
}

func (m *Model) Enable() {
	m.enabled = true
	m.cmds = append(m.cmds, m.spinner.Tick)
}

func (m *Model) Disable() {
	m.enabled = false
}

func (m *Model) Resize() {
	if m.width == 0 || m.height == 0 {
		return
	}
	headerHeight := lipgloss.Height(m.headerView())
	m.viewport.Width = m.width
	height := m.height - m.verticalOffset - headerHeight
	if m.maxHeight > 0 {
		height = min(m.maxHeight, m.height-m.verticalOffset-headerHeight)
	}
	m.viewport.Height = height
	m.viewport.YOffset = headerHeight
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	cmds = append(cmds, m.cmds...)
	m.cmds = nil

	if resize, isResize := msg.(tea.WindowSizeMsg); isResize {
		m.width, m.height = resize.Width, resize.Height
		m.Resize()
	}
	if !m.enabled {
		return m, tea.Batch(cmds...)
	}

	m.viewport.SetContent(m.buffer.String())
	if keyMsg, isKeyMsg := msg.(tea.KeyMsg); isKeyMsg {
		if keyMsg.String() == "f" {
			m.follow = !m.follow
		}
	}

	// follow logic
	if m.follow && m.buffer.Len() != m.lastBufLen {
		m.viewport.GotoBottom()
		m.lastBufLen = m.buffer.Len()
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.enabled {
		return ""
	}
	s := m.headerView()
	if m.buffer.Len() == 0 {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.spinner.View()+" "+styles.InfoText("Waiting for logs").String())
	} else {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.viewport.View())
	}
	return s
}

func (m *Model) headerView() string {
	s := m.followText()
	if m.additionalText != "" {
		s += " | " + m.additionalText
	}
	return lipgloss.NewStyle().
		Padding(0, 1).
		BorderForeground(styles.InfoColor).
		Border(lipgloss.NormalBorder()).
		Width(m.width - (lipgloss.Width(lipgloss.NormalBorder().Left) * 2)).
		Render(s)
}

// followText returns colored info text
// 'follow: %t (press 'f' to start/stop following)\n'
func (m *Model) followText() string {
	b := styles.NewStringBuilder()
	b.WriteInfoColor("follow: ")
	b.WriteSelectColor(strconv.FormatBool(m.follow))
	b.WriteInfoColor(" (press '")
	b.WriteSelectColor("f")
	b.WriteInfoColor("' to start/stop following)")
	return b.String()
}
