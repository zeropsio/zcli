// Package logfile provides an slog.Handler that appends each record to
// a file, opening and closing the file per record.
//
// The handler is intentionally not buffered and does not hold the file
// open between records: partial writes can't corrupt history when zcli
// is killed, multiple concurrent invocations of zcli writing to the
// same file interleave at record granularity, and there is nothing to
// flush on shutdown.
//
// Use it as one of several handlers fed to slog.NewMultiHandler. This
// package does not build the slog.Logger itself — callers compose
// whichever sinks (file, stderr, ...) suit their runtime.
package logfile

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/zeropsio/zcli/src/file"
)

type Config struct {
	Path string
	Mode os.FileMode          // 0 falls back to 0o644
	Opts *slog.HandlerOptions // nil falls back to LevelInfo
}

// New returns an slog.Handler that appends each record to Config.Path.
// It does not retain any open file descriptor; there is nothing to close.
func New(cfg Config) slog.Handler {
	mode := cfg.Mode
	if mode == 0 {
		mode = 0o644
	}
	opts := cfg.Opts
	if opts == nil {
		opts = &slog.HandlerOptions{Level: slog.LevelInfo}
	}
	return &handler{
		path: cfg.Path,
		mode: mode,
		opts: opts,
		mu:   &sync.Mutex{},
	}
}

type handler struct {
	path   string
	mode   os.FileMode
	opts   *slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
	mu     *sync.Mutex
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	f, err := file.Open(h.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, h.mode)
	if err != nil {
		return err
	}
	defer f.Close()

	var th slog.Handler = slog.NewTextHandler(f, h.opts)
	if len(h.attrs) > 0 {
		th = th.WithAttrs(h.attrs)
	}
	for _, g := range h.groups {
		th = th.WithGroup(g)
	}
	return th.Handle(ctx, r)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &clone
}

func (h *handler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.groups = append(append([]string(nil), h.groups...), name)
	return &clone
}
