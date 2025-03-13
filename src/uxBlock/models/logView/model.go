package logView

import (
	"bytes"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type Config struct {
	HorizontalOffset int
}
type Option = generic.Option[Config]

func WithHorizontalOffset(offset int) Option {
	return func(c *Config) {
		c.HorizontalOffset = offset
	}
}

type Model struct {
	config     Config
	buffer     *bytes.Buffer
	viewport   viewport.Model
	spinner    spinner.Model
	lastBufLen int
	follow     bool
}

func New(width, height int, options ...Option) *Model {
	config := generic.ApplyOptions(options...)
	return &Model{
		config:   config,
		buffer:   new(bytes.Buffer),
		viewport: viewport.New(width, height-config.HorizontalOffset),
		spinner:  spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		follow:   true,
	}
}

func (m *Model) Write(p []byte) (n int, err error) {
	return m.buffer.Write(p)
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport.SetContent(m.buffer.String())
	if keyMsg, isKeyMsg := msg.(tea.KeyMsg); isKeyMsg {
		if keyMsg.String() == "f" {
			m.follow = !m.follow
		}
	}
	if m.follow && m.buffer.Len() != m.lastBufLen {
		m.viewport.GotoBottom()
		m.lastBufLen = m.buffer.Len()
	}
	if resize, isResize := msg.(tea.WindowSizeMsg); isResize {
		m.viewport.Width = resize.Width
		m.viewport.Height = resize.Height - m.config.HorizontalOffset
		m.viewport.GotoBottom()
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	s := fmt.Sprintf("follow: %t (press 'f' to start/stop following)\n", m.follow)
	if m.buffer.Len() == 0 {
		s += m.spinner.View() + " " + styles.InfoText("waiting for logs").String() + "\n"
	}
	s += m.viewport.View()
	return s
}
