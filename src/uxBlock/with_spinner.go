package uxBlock

import (
	"context"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock/models/logView"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// SpinnerSpec describes a single spinner managed by WithSpinner.
//
// SuccessMessage / ErrorMessage are the lines that replace the spinner
// when the closure returns. Empty SuccessMessage hides the final line
// on success; empty ErrorMessage falls back to the error's Error()
// rendered with styles.ErrorLine.
type SpinnerSpec struct {
	Message        string
	SuccessMessage string
	ErrorMessage   string
}

// SpinnerHandle is what the closure handed to WithSpinner receives. It
// scopes all spinner mutations to the closure's lifetime: there is no
// way to mutate a spinner after WithSpinner returns.
type SpinnerHandle interface {
	// SetMessage replaces the running spinner's line. Useful for
	// multi-phase work where the prompt should reflect the current phase.
	SetMessage(msg string)

	// LogView returns an io.Writer streaming into the spinner's log
	// pane. Falls back to os.Stdout when not attached to a terminal.
	LogView(opts ...logView.Option) io.Writer
}

// WithSpinner runs fn with a live spinner showing spec.Message. The
// spinner is converted to a final styled line based on the return value
// of fn (and spec.SuccessMessage / spec.ErrorMessage). The closure
// scopes the spinner's lifetime — there is no way for fn to leak the
// handle.
func (b *Blocks) WithSpinner(ctx context.Context, spec SpinnerSpec, fn func(SpinnerHandle) error) error {
	spinner := NewSpinner(styles.NewLine(styles.InfoText(spec.Message)).String())
	stopFunc, send := b.RunSpinners(ctx, []*Spinner{spinner})
	defer stopFunc()

	h := &spinnerHandle{spinner: spinner, send: send}
	err := fn(h)

	switch {
	case err != nil:
		msg := spec.ErrorMessage
		if msg == "" {
			msg = err.Error()
		}
		send(spinner.FinishWithLine(styles.ErrorLine(msg).String()))
	case spec.SuccessMessage != "":
		send(spinner.FinishWithLine(styles.SuccessLine(spec.SuccessMessage).String()))
	default:
		send(spinner.Finish())
	}
	return err
}

type spinnerHandle struct {
	spinner *Spinner
	send    func(tea.Msg)
}

func (h *spinnerHandle) SetMessage(msg string) {
	h.send(spinnerSetMessage{
		id:   h.spinner.id,
		line: styles.NewLine(styles.InfoText(msg)).String(),
	})
}

func (h *spinnerHandle) LogView(opts ...logView.Option) io.Writer {
	if !terminal.IsTerminal() {
		return os.Stdout
	}
	return h.spinner.LogView(opts...)
}

// spinnerSetMessage is the tea.Msg that swaps a running spinner's line.
type spinnerSetMessage struct {
	id   uint64
	line string
}
