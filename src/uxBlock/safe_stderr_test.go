package uxBlock

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeStderr_PassthroughWhenIdle(t *testing.T) {
	var sink bytes.Buffer
	s := NewSafeStderr(&sink)

	n, err := s.Write([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, "hello", sink.String())
}

func TestSafeStderr_BuffersDuringSuspend(t *testing.T) {
	var sink bytes.Buffer
	s := NewSafeStderr(&sink)

	s.suspend()
	_, _ = s.Write([]byte("first "))
	_, _ = s.Write([]byte("second"))
	require.Empty(t, sink.String(), "downstream must remain untouched while suspended")
}

func TestSafeStderr_FlushesOnResume(t *testing.T) {
	var sink bytes.Buffer
	s := NewSafeStderr(&sink)

	s.suspend()
	_, _ = s.Write([]byte("buffered\n"))
	require.Empty(t, sink.String())

	s.resume()
	require.Equal(t, "buffered\n", sink.String(), "resume must flush in order")
}

func TestSafeStderr_NestedSuspendOnlyFlushesAtZero(t *testing.T) {
	// Reference counting: nested suspends don't flush until every
	// matching resume has fired.
	var sink bytes.Buffer
	s := NewSafeStderr(&sink)

	s.suspend()
	s.suspend()
	_, _ = s.Write([]byte("inner"))

	s.resume()
	require.Empty(t, sink.String(), "outer suspend still active")

	s.resume()
	require.Equal(t, "inner", sink.String(), "all suspends resumed -> flush")
}

func TestSafeStderr_BufferRespectsMax(t *testing.T) {
	// The buffer is capped at safeStderrMaxBuffer. Past the cap the
	// excess bytes are dropped from the stderr mirror (still recorded
	// via slog handlers elsewhere in production).
	var sink bytes.Buffer
	s := NewSafeStderr(&sink)
	s.suspend()

	huge := strings.Repeat("x", safeStderrMaxBuffer+1024)
	n, err := s.Write([]byte(huge))
	require.NoError(t, err)
	require.Equal(t, len(huge), n, "Write reports the full length even when buffer is full (caller invariant)")

	s.resume()
	require.Equal(t, safeStderrMaxBuffer, sink.Len(), "flushed bytes capped at safeStderrMaxBuffer")
}

func TestSafeStderr_ConcurrentWritersDoNotRace(t *testing.T) {
	// Race-detector smoke test: many goroutines write while suspend/resume
	// toggles. Compile + run with -race.
	var sink bytes.Buffer
	s := NewSafeStderr(&sink)

	const writers = 16
	const writes = 50
	var wg sync.WaitGroup
	wg.Add(writers)
	for range writers {
		go func() {
			defer wg.Done()
			for range writes {
				_, _ = s.Write([]byte("x"))
			}
		}()
	}
	for range 5 {
		s.suspend()
		s.resume()
	}
	wg.Wait()
}
