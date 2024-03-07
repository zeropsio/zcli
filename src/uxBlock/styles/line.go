package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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
	args := l.args

	for i, arg := range args {
		if typed, ok := arg.(lipgloss.Style); ok {
			if l.styles {
				args[i] = typed.String()
			} else {
				args[i] = typed.Value()
			}
		}
	}
	return fmt.Sprint(args...)
}
