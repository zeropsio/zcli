package models

import (
	"slices"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type TeaModel[T any] interface {
	Update(tea.Msg) (T, tea.Cmd)
}

func Update[T any](sink *CmdSink, msg tea.Msg, model TeaModel[T]) T {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	updated, cmd := model.Update(msg)
	sink.queue = append(sink.queue, cmd)
	return updated
}

type CmdSink struct {
	queue []tea.Cmd
	mu    sync.Mutex
}

func NewCmdSink() *CmdSink {
	return &CmdSink{}
}

func (s *CmdSink) Pour(cmds ...tea.Cmd) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, cmds...)
}

func (s *CmdSink) Filled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queue) > 0
}

func (s *CmdSink) DrainBatch() tea.Cmd {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := slices.Clone(s.queue)
	s.queue = make([]tea.Cmd, 0, len(s.queue))
	return tea.Batch(q...)
}

func (s *CmdSink) DrainSequence() tea.Cmd {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := slices.Clone(s.queue)
	s.queue = make([]tea.Cmd, 0, len(s.queue))
	return tea.Sequence(q...)
}
