package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type StringBuilder struct {
	*strings.Builder
}

func NewStringBuilder() *StringBuilder {
	return &StringBuilder{
		Builder: new(strings.Builder),
	}
}

func (b *StringBuilder) WriteInfoColor(s string) {
	b.WriteString(infoStyle.Render(s))
}

func (b *StringBuilder) WriteSuccessColor(s string) {
	b.WriteString(successStyle.Render(s))
}

func (b *StringBuilder) WriteSelectColor(s string) {
	b.WriteString(selectStyle.Render(s))
}

func (b *StringBuilder) WriteWarningColor(s string) {
	b.WriteString(warningStyle.Render(s))
}

func (b *StringBuilder) WriteErrorColor(s string) {
	b.WriteString(errorStyle.Render(s))
}

func (b *StringBuilder) WriteStyledColor(style lipgloss.Style, s string) {
	b.WriteString(style.Inline(true).Render(s))
}
