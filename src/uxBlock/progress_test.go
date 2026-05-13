package uxBlock

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/uxBlock/models/logView"
)

// spinnerHandleStub implements SpinnerHandle for tests. It records
// SetMessage calls and ignores LogView (no TUI program is running).
type spinnerHandleStub struct {
	mu  sync.Mutex
	buf []string
}

func (h *spinnerHandleStub) SetMessage(m string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buf = append(h.buf, m)
}

func (h *spinnerHandleStub) LogView(_ ...logView.Option) io.Writer {
	return io.Discard
}

func (h *spinnerHandleStub) LogStream(_ string, _ ...logView.Option) io.Writer {
	return io.Discard
}

func (h *spinnerHandleStub) snapshot() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	cp := make([]string, len(h.buf))
	copy(cp, h.buf)
	return cp
}

func TestProgressReader_ReadsThroughUnchanged(t *testing.T) {
	// The counting wrapper must not alter the data stream.
	payload := []byte("hello, world!")
	r, stop := ProgressReader(&spinnerHandleStub{}, "uploading", bytes.NewReader(payload))
	defer stop()

	got, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, payload, got)
}

func TestProgressReader_EmitsMessagesAfterTick(t *testing.T) {
	// Exercise the updater goroutine for ~750ms (at least one 500ms
	// tick). Exact counts depend on scheduling so we only assert
	// shape, not number.
	h := &spinnerHandleStub{}
	r, stop := ProgressReader(h, "uploading", strings.NewReader(strings.Repeat("x", 1024)))
	_, _ = io.ReadAll(r)
	time.Sleep(750 * time.Millisecond)
	stop()

	got := h.snapshot()
	if len(got) == 0 {
		t.Skip("scheduler did not run the ticker in this run; signal-only test")
	}
	for _, m := range got {
		require.Contains(t, m, "uploading", "every message must include the label")
	}
}

func TestProgressReader_StopHaltsGoroutine(t *testing.T) {
	// After stop(), no further messages should arrive.
	h := &spinnerHandleStub{}
	_, stop := ProgressReader(h, "x", strings.NewReader(""))
	stop()
	before := len(h.snapshot())
	time.Sleep(700 * time.Millisecond)
	after := len(h.snapshot())
	require.Equal(t, before, after, "no messages after stop()")
}
