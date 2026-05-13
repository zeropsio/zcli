package uxBlock

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// Print emits each line of msg styled via the given style fn to stderr.
// Multi-line input is split on \n so every line receives the style
// decoration. Print has no logging side effect — callers wanting trace
// data write to slog directly.
func (b *Blocks) Print(style func(string) styles.Line, msg string) {
	scan := bufio.NewScanner(strings.NewReader(msg))
	for scan.Scan() {
		fmt.Fprintln(os.Stderr, style(scan.Text()).String())
	}
}

// Printf is fmt.Sprintf + Print.
func (b *Blocks) Printf(style func(string) styles.Line, format string, args ...any) {
	b.Print(style, fmt.Sprintf(format, args...))
}
