package uxBlock

import (
	"context"
	"sync"

	bubblesSpinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (b *uxBlocks) RunSpinners(ctx context.Context, spinners []*Spinner, auxOptions ...SpinnerOption) func() {
	cfg := spinnerConfig{}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	if !b.isTerminal {
		return func() {
			for _, spinner := range spinners {
				b.PrintLine(spinner.text)
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
		//nolint:errcheck
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

func XXX(cmdList ...tea.Cmd) func() tea.Msg {
	return func() tea.Msg {
		return MergeMessage(cmdList)
	}
}
func (m *spinnerModel) Init() tea.Cmd {
	ticks := make([]tea.Cmd, len(m.spinners))
	for i := range m.spinners {
		ticks[i] = m.spinners[i].init()
	}

	return XXX(ticks...)
}

func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerEndCmd:
		m.quiting = true
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
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
		return m, XXX(cmdList...)
	}
	return m, nil
}

func (m *spinnerModel) View() string {
	var s string
	for _, spinner := range m.spinners {
		if m.canceled {
			s += "canceled\n"
		} else {
			s += spinner.view() + spinner.text + "\n"
		}
	}

	return s
}

type spinnerConfig struct {
}

type SpinnerOption = func(cfg *spinnerConfig)

type Spinner struct {
	text     string
	finished bool
	spinner  bubblesSpinner.Model
}

func NewSpinner(text string) *Spinner {
	return &Spinner{
		text:    text,
		spinner: bubblesSpinner.New(bubblesSpinner.WithSpinner(bubblesSpinner.MiniDot)),
	}
}

func (s *Spinner) SetMessage(text string) *Spinner {
	s.text = text

	return s
}

func (s *Spinner) Finish(text string) *Spinner {
	s.text = text
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
