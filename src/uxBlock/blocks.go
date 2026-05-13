// Package uxBlock provides building blocks for UX and communication with a user.
package uxBlock

import (
	"context"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

//go:generate go run --mod=mod github.com/golang/mock/mockgen -source=$GOFILE -destination=$PWD/mocks/$GOFILE -package=mocks

type UxBlocks interface {
	Print(style func(string) styles.Line, msg string)
	Printf(style func(string) styles.Line, format string, args ...any)
	RunSpinners(ctx context.Context, spinners []*Spinner) (func(), func(msg tea.Msg))
}

type Blocks struct {
	isTerminal     bool
	TerminalWidth  int
	TerminalHeight int

	// header is a status line rendered above every active spinner —
	// typically the resolved scope (project, service, region) so the
	// user always sees what they're operating on. Empty by default.
	header string

	// stderr wraps the process stderr. It suspends raw writes while a
	// Bubble Tea program is drawing, then flushes after the program
	// stops, so subprocess noise or slog stderr records never tear
	// the TUI frame.
	stderr *SafeStderr

	// ctxCancel is used to cancel the context of the command.
	// Bubbles package that we use to render graphic components steals the signal handler.
	// In case that I want to cancel the context of a running component, e.g. spinner, the main context is not canceled.
	// Therefore, we need to pass the cancel function to the uxBlocks.
	// If the graphic component is canceled, we cancel the main context.
	ctxCancel context.CancelFunc
}

// Stderr returns the TUI-aware writer that callers should use anywhere
// they would otherwise write to os.Stderr (the slog stderr handler,
// subprocess Cmd.Stderr, early-init warnings).
func (b *Blocks) Stderr() io.Writer { return b.stderr }

// SetHeader replaces the persistent status line shown above every
// active spinner. Pass an empty string to clear it.
func (b *Blocks) SetHeader(s string) {
	b.header = s
}

func NewBlocks(
	isTerminal bool,
	terminalWidth int,
	terminalHeight int,
	ctxCancel context.CancelFunc,
) *Blocks {
	// safety check
	if ctxCancel == nil {
		ctxCancel = func() {}
	}

	return &Blocks{
		isTerminal:     isTerminal,
		TerminalWidth:  terminalWidth,
		TerminalHeight: terminalHeight,
		stderr:         NewSafeStderr(os.Stderr),
		ctxCancel:      ctxCancel,
	}
}
