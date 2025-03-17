package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/terminal"
)

const (
	NewLine   = "\n"
	EmptyLine = ""
)

type Printer struct {
	out io.Writer
}

func NewPrinter(out io.Writer) Printer {
	return Printer{
		out: out,
	}
}

func (p Printer) Printf(format string, args ...any) {
	fmt.Fprintf(p.out, format, args...)
}

func (p Printer) Print(a ...any) {
	fmt.Fprint(p.out, a...)
}

func (p Printer) Println(a ...any) {
	fmt.Fprintln(p.out, a...)
}

func (p Printer) PrintLines(lines ...string) {
	p.Println(strings.Join(lines, NewLine))
}

func Style(s lipgloss.Style, text string) string {
	if !terminal.IsTerminal() {
		return s.Value()
	}
	return s.Render(text)
}

func (p *Printer) GetWriter() io.Writer {
	return p.out
}
