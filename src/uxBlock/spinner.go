package uxBlock

import (
	"context"
	"io"
	"sync"

	bubblesSpinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/uxBlock/models/logView"
)

func (b *Blocks) RunSpinners(ctx context.Context, spinners []*Spinner) (func(), func(msg tea.Msg)) {
	if !b.isTerminal {
		return func() {
				for _, spinner := range spinners {
					b.PrintInfoText(spinner.line)
				}
			},
			func(tea.Msg) {}
	}

	model := &spinnerModel{
		uxBlocks: b,
		spinners: spinners,
	}

	p := tea.NewProgram(model, tea.WithoutSignalHandler(), tea.WithContext(ctx), tea.WithFPS(100))
	go func() {
		//nolint:errcheck // Why: I'm not interest in the error
		p.Run()
		if model.canceled {
			b.ctxCancel()
		}
	}()

	return func() {
			p.Send(spinnerEndCmd{})
			p.Wait()
		}, func(msg tea.Msg) {
			p.Send(msg)
		}
}

type spinnerEndCmd struct {
}

type spinnerModel struct {
	spinners []*Spinner
	uxBlocks *Blocks

	quiting  bool
	canceled bool
}

func (m *spinnerModel) Init() tea.Cmd {
	ticks := make([]tea.Cmd, len(m.spinners))
	for i := range m.spinners {
		ticks[i] = m.spinners[i].init()
	}
	return tea.Batch(ticks...)
}

func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quiting {
		return m, tea.Quit
	}

	cmds := make([]tea.Cmd, 0, len(m.spinners))
	for _, s := range m.spinners {
		cmds = append(cmds, s.update(msg))
	}

	switch msg := msg.(type) {
	case spinnerEndCmd:
		m.quiting = true
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.canceled = true
			m.quiting = true
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *spinnerModel) View() string {
	var s string
	for _, spinner := range m.spinners {
		if m.canceled {
			s += "canceled\n"
		} else {
			s += spinner.view() + "\n"
		}
	}
	return s
}

// Internal ID management. Used during animating to ensure that frame messages
// are received only by spinner components that sent them.
var (
	lastID uint64
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() uint64 {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

type Spinner struct {
	id       uint64
	line     string
	finished bool
	spinner  bubblesSpinner.Model
	logView  *logView.Model
}

func NewSpinner(line string) *Spinner {
	return &Spinner{
		id:      nextID(),
		line:    line,
		spinner: bubblesSpinner.New(bubblesSpinner.WithSpinner(bubblesSpinner.MiniDot)),
		logView: logView.New(),
	}
}

func (s *Spinner) LogView(opts ...logView.Option) io.WriteCloser {
	s.logView = gn.ApplyOptionsWithDefault(*s.logView, opts...)
	s.logView.Enable()
	r, w := io.Pipe()
	go func() {
		defer r.Close()
		_, err := io.Copy(s.logView, r)
		if err != nil {
			panic(err)
		}
	}()
	return w
}

func (s *Spinner) init() tea.Cmd {
	return tea.Sequence(
		s.spinner.Tick,
		s.logView.Init(),
	)
}

func (s *Spinner) update(msg tea.Msg) (cmd tea.Cmd) {
	var cmds []tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	cmds = append(cmds, cmd)
	s.logView, cmd = s.logView.Update(msg)
	cmds = append(cmds, cmd)

	if finishMsg, isFinnish := msg.(spinnerFinish); isFinnish && finishMsg.id == s.id {
		if finishMsg.msg != nil {
			s.line = *finishMsg.msg
		}
		s.finished = true
	}

	return tea.Batch(cmds...)
}

func (s *Spinner) view() string {
	if s.finished {
		if s.logView.Enabled() {
			return lipgloss.JoinVertical(
				lipgloss.Left,
				s.logView.View(),
				s.line,
			)
		}
		return s.line
	}
	if s.logView.Enabled() {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			s.logView.View(),
			s.spinner.View()+" "+s.line,
		)
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		s.spinner.View()+" "+s.line,
	)
}

type spinnerFinish struct {
	id  uint64
	msg *string
}

func (s *Spinner) Finish() tea.Msg {
	return spinnerFinish{
		id: s.id,
	}
}

func (s *Spinner) FinishWithLine(msg string) tea.Msg {
	return spinnerFinish{
		id:  s.id,
		msg: &msg,
	}
}
