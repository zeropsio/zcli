package uxBlock

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// styledOutput receives a styles.Line and renders it to stderr followed by a
// newline. Plain io.Writer semantics — no logging side effects.
func (b *Blocks) styledOutput(line styles.Line) {
	_, _ = fmt.Fprintln(os.Stderr, line.String())
}

// mirrorToLog dispatches a styled line to slog.Default at the given level
// with the styles stripped, so the debug log records what the user saw.
func mirrorToLog(level slog.Level, line styles.Line) {
	slog.Default().Log(context.Background(), level, line.DisableStyle().String())
}

func (b *Blocks) PrintSuccessText(in string) {
	scan := bufio.NewScanner(bytes.NewBufferString(in))
	for scan.Scan() {
		b.PrintInfo(styles.SuccessLine(scan.Text()))
	}
}

func (b *Blocks) PrintSuccessTextf(format string, args ...any) {
	b.PrintSuccessText(fmt.Sprintf(format, args...))
}

func (b *Blocks) PrintInfoText(in string) {
	scan := bufio.NewScanner(bytes.NewBufferString(in))
	for scan.Scan() {
		b.PrintInfo(styles.InfoLine(scan.Text()))
	}
}

func (b *Blocks) PrintInfoTextf(format string, args ...any) {
	b.PrintInfoText(fmt.Sprintf(format, args...))
}

func (b *Blocks) PrintInfo(line styles.Line) {
	b.styledOutput(line)
	mirrorToLog(slog.LevelInfo, line)
}

func (b *Blocks) PrintWarningText(in string) {
	scan := bufio.NewScanner(bytes.NewBufferString(in))
	for scan.Scan() {
		b.PrintWarning(styles.WarningLine(scan.Text()))
	}
}

func (b *Blocks) PrintWarningTextf(format string, args ...any) {
	b.PrintWarningText(fmt.Sprintf(format, args...))
}

func (b *Blocks) PrintWarning(line styles.Line) {
	b.styledOutput(line)
	mirrorToLog(slog.LevelWarn, line)
}

func (b *Blocks) PrintErrorText(in string) {
	scan := bufio.NewScanner(bytes.NewBufferString(in))
	for scan.Scan() {
		b.PrintError(styles.ErrorLine(scan.Text()))
	}
}

func (b *Blocks) PrintErrorTextf(format string, args ...any) {
	b.PrintErrorText(fmt.Sprintf(format, args...))
}

func (b *Blocks) PrintError(line styles.Line) {
	b.styledOutput(line)
	mirrorToLog(slog.LevelError, line)
}
