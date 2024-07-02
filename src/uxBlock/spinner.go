package uxBlock

import (
	"context"
	"sync"

	bubblesSpinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func (b *uxBlocks) RunSpinners(ctx context.Context, spinners []*Spinner, auxOptions ...SpinnerOption) func() {
	cfg := spinnerConfig{}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	if !b.isTerminal {
		return func() {
			for _, spinner := range spinners {
				b.PrintInfo(spinner.line)
			}
		}
	}

	model := &spinnerModel{
		cfg:      cfg,
		uxBlocks: b,
		spinners: spinners,
	}

	p := tea.NewProgram(model, tea.WithoutSignalHandler(), tea.WithContext(ctx))
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
	cfg      spinnerConfig
	spinners []*Spinner
	uxBlocks *uxBlocks

	quiting  bool
	canceled bool
}

type MergeMessage []tea.Cmd

func sequence(cmdList ...tea.Cmd) func() tea.Msg {
	return func() tea.Msg {
		return MergeMessage(cmdList)
	}
}
func (m *spinnerModel) Init() tea.Cmd {
	ticks := make([]tea.Cmd, len(m.spinners))
	for i := range m.spinners {
		ticks[i] = m.spinners[i].init()
	}

	return sequence(ticks...)
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

	case MergeMessage:
		cmdList := make([]tea.Cmd, len(msg))

		var lock sync.Mutex

		wg := sync.WaitGroup{}
		for i := range msg {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				var teaMsg tea.Msg
				if msg[i] != nil {
					teaMsg = msg[i]()
				}

				cmd := m.spinners[i].update(teaMsg)

				lock.Lock()
				cmdList[i] = cmd
				lock.Unlock()
			}(i)
		}
		wg.Wait()
		return m, sequence(cmdList...)
	}
	return m, nil
}

func (m *spinnerModel) View() string {
	var s string
	for _, spinner := range m.spinners {
		if m.canceled {
			s += "canceled\n"
		} else {
			line := spinner.line.String()
			if line != "" {
				s += spinner.view() + spinner.line.String() + "\n"
			}
		}
	}

	return s
}

type spinnerConfig struct {
}

type SpinnerOption = func(cfg *spinnerConfig)

type Spinner struct {
	line     styles.Line
	finished bool
	spinner  bubblesSpinner.Model
}

func NewSpinner(line styles.Line) *Spinner {
	return &Spinner{
		line:    line,
		spinner: bubblesSpinner.New(bubblesSpinner.WithSpinner(bubblesSpinner.MiniDot)),
	}
}

func (s *Spinner) SetMessage(text styles.Line) *Spinner {
	s.line = text

	return s
}

func (s *Spinner) Finish(text styles.Line) *Spinner {
	s.line = text
	s.finished = true

	return s
}

func (s *Spinner) init() func() tea.Msg {
	return s.spinner.Tick
}

func (s *Spinner) update(msg tea.Msg) (cmd tea.Cmd) {
	if !s.finished {
		s.spinner, cmd = s.spinner.Update(msg)
	}
	return
}

func (s *Spinner) view() string {
	if !s.finished {
		return s.spinner.View() + " "
	}

	return ""
}
