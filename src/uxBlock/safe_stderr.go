package uxBlock

import (
	"io"
	"sync"
)

// SafeStderr is a TUI-aware io.Writer wrapping a downstream writer
// (usually os.Stderr). While a Bubble Tea program is active, writes
// are buffered instead of forwarded — preventing raw bytes from
// racing the program's draw loop and tearing the frame. When the
// program stops, the buffer is flushed to the downstream writer in
// order.
//
// Callers should obtain the SafeStderr exclusively via (*Blocks).Stderr()
// and use it everywhere they would otherwise write to os.Stderr:
//   - the slog stderr handler
//   - subprocess Stderr destinations
//   - direct fmt.Fprintln() for early-init warnings
//
// Anything that bypasses SafeStderr risks clobbering the TUI.
type SafeStderr struct {
	downstream io.Writer
	mu         sync.Mutex
	teaActive  int
	buffer     []byte
}

// safeStderrMaxBuffer caps the held-back buffer at 1 MiB so a flood of
// subprocess output during a long spinner can't OOM the CLI. Past the
// cap, additional writes are silently dropped from the stderr mirror
// (they still land in the log file via the slog file handler).
const safeStderrMaxBuffer = 1 << 20

// NewSafeStderr wraps downstream. Pass os.Stderr in production and a
// *bytes.Buffer in tests.
func NewSafeStderr(downstream io.Writer) *SafeStderr {
	return &SafeStderr{downstream: downstream}
}

// Write forwards to downstream when no Bubble Tea program is active.
// While one is active, the bytes are appended to an internal buffer
// (up to safeStderrMaxBuffer) and flushed on the next deactivation.
func (s *SafeStderr) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.teaActive > 0 {
		room := safeStderrMaxBuffer - len(s.buffer)
		if room <= 0 {
			return len(p), nil
		}
		if len(p) > room {
			s.buffer = append(s.buffer, p[:room]...)
		} else {
			s.buffer = append(s.buffer, p...)
		}
		return len(p), nil
	}
	return s.downstream.Write(p)
}

// suspend marks a new Bubble Tea program as active. Reference-counted
// so nested programs (we don't currently have any, but future-proof)
// don't flush early.
func (s *SafeStderr) suspend() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teaActive++
}

// resume decrements the active-program count. When it reaches zero,
// any buffered bytes are flushed downstream in order.
func (s *SafeStderr) resume() {
	s.mu.Lock()
	buf := s.buffer
	s.teaActive--
	if s.teaActive > 0 || len(buf) == 0 {
		s.mu.Unlock()
		return
	}
	s.buffer = nil
	s.mu.Unlock()
	// Write outside the lock to avoid holding it across a syscall.
	_, _ = s.downstream.Write(buf)
}
