package uxBlock

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

const (
	SuccessIcon   = "✅ "
	ErrorIcon     = "❌ "
	SelectionIcon = "❔ "
	InfoIcon      = "ℹ️"
	WarningIcon   = "⚠️"
)

var defaultStyle = lipgloss.NewStyle().
	Renderer(lipgloss.NewRenderer(os.Stdout))

var successColor = defaultStyle.
	Foreground(lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{TrueColor: "#00ff5f", ANSI256: "47", ANSI: "0"},
		Dark:  lipgloss.CompleteColor{TrueColor: "#00ff5f", ANSI256: "47", ANSI: "0"},
	}).
	Bold(false)

var errorColor = defaultStyle.
	Foreground(lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{TrueColor: "#FF000", ANSI256: "196", ANSI: "0"},
		Dark:  lipgloss.CompleteColor{TrueColor: "#FF000", ANSI256: "196", ANSI: "0"},
	}).
	Bold(true)

var warningColor = defaultStyle.
	Foreground(lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{TrueColor: "#ffff87", ANSI256: "228", ANSI: "0"},
		Dark:  lipgloss.CompleteColor{TrueColor: "#ffff87", ANSI256: "228", ANSI: "0"},
	}).
	Bold(true)

var infoColor = defaultStyle.
	Foreground(lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{TrueColor: "#00afff", ANSI256: "039", ANSI: "0"},
		Dark:  lipgloss.CompleteColor{TrueColor: "#00afff", ANSI256: "039", ANSI: "0"},
	}).
	Bold(true)

func ErrorText(text string) lipgloss.Style {
	return errorColor.Copy().SetString(text)
}

func SuccessText(text string) lipgloss.Style {
	return successColor.Copy().SetString(text)
}

func WarningText(text string) lipgloss.Style {
	return warningColor.Copy().SetString(text)
}

func InfoText(text string) lipgloss.Style {
	return infoColor.Copy().SetString(text)
}

type line struct {
	args   []interface{}
	styles bool
}

func NewLine(args ...interface{}) line {
	return line{
		args:   args,
		styles: true,
	}
}

func (l line) DisableStyle() line {
	l.styles = false
	return l
}

func (l line) Merge(lines ...line) line {
	for _, line := range lines {
		l.args = append(l.args, line.args...)
	}
	return l
}

func (l line) NotEmpty() bool {
	return len(l.args) > 0
}

func (l line) String() string {
	if l.styles {
		return fmt.Sprint(l.args...)
	}
	return fmt.Sprint(removeStyles(l.args)...)
}

func removeStyles(args []interface{}) []interface{} {
	for i, arg := range args {
		if typed, ok := arg.(lipgloss.Style); ok {
			args[i] = typed.Value()
		}
	}

	return args
}
