// Package logger configures the structured logger used by zcli.
//
// There is exactly one structured logger: the *slog.Logger installed as
// slog.Default during startup. Callers should use the standard library
// log/slog package directly (slog.Info, slog.Debug, slog.LogAttrs, ...).
//
// Two sinks are wired:
//   - file:   always on when Config.FilePath is set; text-formatted,
//     written via VarLogHook-style append-on-each-record so the
//     file remains valid even if the process is killed.
//   - stderr: only on when Config.Verbose is true.
//
// User-facing terminal output (styled lines, prompts, spinners) is NOT
// the responsibility of this package — see src/uxBlock.
package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/zeropsio/zcli/src/file"
)

type Config struct {
	FilePath string      // empty disables the file sink
	FileMode os.FileMode // 0 falls back to 0o644
	Verbose  bool        // enables the stderr sink and lowers the level to Debug
}

// New constructs the slog.Logger described by cfg. The returned io.Closer
// flushes/closes any owned sinks; callers should defer Close on exit.
//
// New does not install the logger as slog.Default; callers do that
// explicitly so test code can use isolated loggers.
func New(cfg Config) (*slog.Logger, io.Closer, error) {
	level := slog.LevelInfo
	if cfg.Verbose {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}

	var handlers []slog.Handler
	var closers []io.Closer

	if cfg.FilePath != "" {
		fileMode := cfg.FileMode
		if fileMode == 0 {
			fileMode = 0o644
		}
		h := &appendFileHandler{
			path:     cfg.FilePath,
			fileMode: fileMode,
			opts:     opts,
			mu:       &sync.Mutex{},
		}
		handlers = append(handlers, h)
	}

	if cfg.Verbose {
		handlers = append(handlers, slog.NewTextHandler(os.Stderr, opts))
	}

	var handler slog.Handler
	switch len(handlers) {
	case 0:
		handler = slog.NewTextHandler(io.Discard, opts)
	case 1:
		handler = handlers[0]
	default:
		handler = multiHandler(handlers)
	}

	return slog.New(handler), multiCloser(closers), nil
}

// multiHandler fans an slog.Record out to every child handler. The first
// non-nil error from a child is returned but every child is still given a
// chance to handle the record (matches logrus hook behavior).
type multiHandler []slog.Handler

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make(multiHandler, len(m))
	for i, h := range m {
		out[i] = h.WithAttrs(attrs)
	}
	return out
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	out := make(multiHandler, len(m))
	for i, h := range m {
		out[i] = h.WithGroup(name)
	}
	return out
}

// appendFileHandler opens the configured log file on every Handle call,
// appends one record (text format), and closes the file. This matches the
// historical logrus VarLogHook behavior: durable across crashes, robust
// against concurrent invocations of zcli writing to the same file, and
// trivially safe under signals.
type appendFileHandler struct {
	path     string
	fileMode os.FileMode
	opts     *slog.HandlerOptions
	attrs    []slog.Attr
	groups   []string
	mu       *sync.Mutex
}

func (h *appendFileHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *appendFileHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	f, err := file.Open(h.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, h.fileMode)
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

func (h *appendFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &clone
}

func (h *appendFileHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.groups = append(append([]string(nil), h.groups...), name)
	return &clone
}

type multiCloser []io.Closer

func (m multiCloser) Close() error {
	var errs []error
	for _, c := range m {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
