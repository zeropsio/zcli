package uxBlock

import (
	"context"
	"io"

	bubblesSpinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/uxBlock/models/logView"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func (b *Blocks) RunSpinners(ctx context.Context, spinners []*Spinner) func() {
	if !b.isTerminal {
		return func() {
			for _, spinner := range spinners {
				b.PrintInfo(spinner.line)
			}
		}
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
	switch msg := msg.(type) {
	case spinnerEndCmd:
		m.quiting = true
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.canceled = true
			m.quiting = true
			return m, tea.Quit
		}
	}
	cmds := make([]tea.Cmd, 0, len(m.spinners))
	for _, s := range m.spinners {
		cmds = append(cmds, s.update(msg))
	}
	return m, tea.Batch(cmds...)
}

func (m *spinnerModel) View() string {
	var s string
	for _, spinner := range m.spinners {
		if spinner.finished && !spinner.endedWithError {
			continue
		}
		if m.canceled {
			s += "canceled\n"
		} else {
			s += spinner.view() + "\n"
		}
	}

	return s
}

type Spinner struct {
	line           styles.Line
	finished       bool
	endedWithError bool
	spinner        bubblesSpinner.Model
	logView        *logView.Model
}

func NewSpinner(line styles.Line, width, height int) *Spinner {
	return &Spinner{
		line:    line,
		spinner: bubblesSpinner.New(bubblesSpinner.WithSpinner(bubblesSpinner.MiniDot)),
		logView: logView.New(
			width,
			height,
			logView.WithVerticalOffset(2),
		),
	}
}

func (s *Spinner) SetMessage(text styles.Line) *Spinner {
	s.line = text

	return s
}

func (s *Spinner) LogView() io.WriteCloser {
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

func (s *Spinner) FinishEmpty() *Spinner {
	s.finished = true
	return s
}

func (s *Spinner) Finish(text styles.Line) *Spinner {
	s.line = text
	s.finished = true
	return s
}

func (s *Spinner) FinishWithError(text styles.Line) *Spinner {
	s.line = text
	s.finished = true
	s.endedWithError = true
	return s
}

func (s *Spinner) init() tea.Cmd {
	return tea.Sequence(
		s.spinner.Tick,
		s.logView.Init(),
	)
}

func (s *Spinner) update(msg tea.Msg) (cmd tea.Cmd) {
	var cmds []tea.Cmd
	if s.finished {
		return nil
	}
	s.spinner, cmd = s.spinner.Update(msg)
	cmds = append(cmds, cmd)
	s.logView, cmd = s.logView.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (s *Spinner) view() string {
	var l string
	if s.finished {
		if s.endedWithError && s.logView.Enabled {
			l += s.logView.View() + "\n"
			return l + s.line.String()
		}
		return s.line.String()
	}
	if s.logView.Enabled {
		l += s.logView.View() + "\n"
	}
	return l + s.spinner.View() + " " + s.line.String()
}
