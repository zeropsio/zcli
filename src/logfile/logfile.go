// Package logfile provides an slog.Handler that appends each record to
// a file, opening and closing the file per record.
//
// The handler is intentionally not buffered and does not hold the file
// open between records: partial writes can't corrupt history when zcli
// is killed, multiple concurrent invocations of zcli writing to the
// same file interleave at record granularity, and there is nothing to
// flush on shutdown.
//
// The file is capped at Config.MaxBytes. When a write would push the
// file past the cap, the current file is rotated to "<path>.1"
// (overwriting any prior .1) and a fresh file is started. At most one
// rotated file is kept; total on-disk footprint is bounded at
// ~2 * MaxBytes.
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

// DefaultMaxBytes caps the live log file. Pick a size big enough to
// hold a few minutes of verbose Debug output while keeping the rotated
// pair below 10 MB total.
const DefaultMaxBytes = 4 * 1024 * 1024

type Config struct {
	Path     string
	Mode     os.FileMode          // 0 falls back to 0o644
	Opts     *slog.HandlerOptions // nil falls back to LevelInfo
	MaxBytes int64                // 0 falls back to DefaultMaxBytes; negative disables the cap
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
	maxBytes := cfg.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	return &handler{
		path:     cfg.Path,
		mode:     mode,
		opts:     opts,
		maxBytes: maxBytes,
		mu:       &sync.Mutex{},
	}
}

type handler struct {
	path     string
	mode     os.FileMode
	opts     *slog.HandlerOptions
	maxBytes int64
	attrs    []slog.Attr
	groups   []string
	mu       *sync.Mutex
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.maxBytes > 0 {
		if info, err := os.Stat(h.path); err == nil && info.Size() >= h.maxBytes {
			// Rotate: rename to .1 (overwriting any prior rotation) and
			// start fresh. Ignore errors — if rotation fails, the next
			// write will retry and at worst the file grows past the cap
			// until disk pressure surfaces the real problem.
			_ = os.Rename(h.path, h.path+".1")
		}
	}

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
