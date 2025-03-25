package styles

import (
	"strings"
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
	b.WriteString(infoColor.Render(s))
}

func (b *StringBuilder) WriteSuccessColor(s string) {
	b.WriteString(successColor.Render(s))
}

func (b *StringBuilder) WriteSelectColor(s string) {
	b.WriteString(selectColor.Render(s))
}

func (b *StringBuilder) WriteWarningColor(s string) {
	b.WriteString(warningColor.Render(s))
}

func (b *StringBuilder) WriteErrorColor(s string) {
	b.WriteString(errorColor.Render(s))
}
