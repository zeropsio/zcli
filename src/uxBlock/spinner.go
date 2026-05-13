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

// runSpinners is the low-level Bubble Tea program launcher used by
// WithSpinner and WithSpinners. Returns a stop closure that ends the
// program and a send closure that posts tea.Msg values (typically
// spinnerFinish / spinnerSetMessage). Not exported: callers should
// use the closure-scoped WithSpinner / WithSpinners APIs.
func (b *Blocks) runSpinners(ctx context.Context, spinners []*Spinner) (func(), func(msg tea.Msg)) {
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

	// While the Bubble Tea program is drawing, suspend the SafeStderr so
	// any raw write (slog stderr handler, subprocess stderr that slipped
	// past the slog adapter) is buffered instead of tearing the frame.
	// Resumed after the program exits.
	b.stderr.suspend()

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
			b.stderr.resume()
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
	if m.uxBlocks.header != "" {
		s += m.uxBlocks.header + "\n"
	}
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
	logView  *logView.Model // default unnamed pane (preserved for LogView)

	// namedLogs holds additional log panes addressable by name. Each
	// pane renders independently in insertion order, stacked above the
	// spinner line. Used by callers (e.g. servicePush) that interleave
	// concurrent streams (build/prepare) and want them visually
	// separated rather than merged into one pane.
	namedLogs  map[string]*logView.Model
	namedOrder []string
	namedMu    sync.Mutex
}

func NewSpinner(line string) *Spinner {
	return &Spinner{
		id:        nextID(),
		line:      line,
		spinner:   bubblesSpinner.New(bubblesSpinner.WithSpinner(bubblesSpinner.MiniDot)),
		logView:   logView.New(),
		namedLogs: map[string]*logView.Model{},
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

// LogStream returns a writer streaming into a named log pane on this
// spinner. The first call with a given name creates the pane (which
// renders above the spinner line, stacked under any earlier streams).
// Subsequent calls with the same name reuse it. Names are
// case-sensitive; pick short labels — they're not rendered, just used
// to address the pane.
func (s *Spinner) LogStream(name string, opts ...logView.Option) io.WriteCloser {
	s.namedMu.Lock()
	lv, exists := s.namedLogs[name]
	if !exists {
		lv = gn.ApplyOptionsWithDefault(*logView.New(), opts...)
		lv.Enable()
		s.namedLogs[name] = lv
		s.namedOrder = append(s.namedOrder, name)
	}
	s.namedMu.Unlock()
	r, w := io.Pipe()
	go func() {
		defer r.Close()
		_, err := io.Copy(lv, r)
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

	// Forward msg to every named log pane so its scroll/follow state
	// keeps updating. The map is mutated only via LogStream — take a
	// snapshot under the lock and update outside it.
	s.namedMu.Lock()
	views := make([]*logView.Model, 0, len(s.namedOrder))
	for _, name := range s.namedOrder {
		views = append(views, s.namedLogs[name])
	}
	s.namedMu.Unlock()
	for i, lv := range views {
		updated, c := lv.Update(msg)
		cmds = append(cmds, c)
		s.namedMu.Lock()
		s.namedLogs[s.namedOrder[i]] = updated
		s.namedMu.Unlock()
	}

	if finishMsg, isFinnish := msg.(spinnerFinish); isFinnish && finishMsg.id == s.id {
		if finishMsg.msg != nil {
			s.line = *finishMsg.msg
		}
		s.finished = true
	}

	if setMsg, ok := msg.(spinnerSetMessage); ok && setMsg.id == s.id && !s.finished {
		s.line = setMsg.line
	}

	return tea.Batch(cmds...)
}

func (s *Spinner) view() string {
	parts := make([]string, 0, 4)

	// Named log panes stack above the default pane, in insertion order.
	s.namedMu.Lock()
	for _, name := range s.namedOrder {
		if lv := s.namedLogs[name]; lv != nil && lv.Enabled() {
			parts = append(parts, lv.View())
		}
	}
	s.namedMu.Unlock()

	if s.logView.Enabled() {
		parts = append(parts, s.logView.View())
	}

	if s.finished {
		parts = append(parts, s.line)
	} else {
		parts = append(parts, s.spinner.View()+" "+s.line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
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
