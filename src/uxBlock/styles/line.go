package styles

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/zeropsio/zcli/src/terminal"
)

type Line struct {
	args   []interface{}
	styles bool
}

func NewLine(args ...interface{}) Line {
	return Line{
		args:   args,
		styles: true,
	}
}

func (l Line) DisableStyle() Line {
	l.styles = false
	return l
}

func (l Line) Merge(lines ...Line) Line {
	for _, line := range lines {
		l.args = append(l.args, line.args...)
	}
	return l
}

func (l Line) NotEmpty() bool {
	return len(l.args) > 0
}

func (l Line) String() string {
	args := slices.Clone(l.args)

	for i, arg := range args {
		if typed, ok := arg.(lipgloss.Style); ok {
			if l.styles && terminal.IsTerminal() {
				args[i] = typed.String()
			} else {
				args[i] = typed.Value()
			}
		}
	}
	return fmt.Sprint(args...)
}
