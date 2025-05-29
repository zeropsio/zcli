package terminal

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/zeropsio/zcli/src/constants"
	"golang.org/x/term"
)

type Mode string

const (
	ModeAuto     Mode = "auto"
	ModeDisabled Mode = "disabled"
	ModeEnabled  Mode = "enabled"
)

func (m Mode) IsAuto(other Mode) bool {
	return ModeAuto == other
}
func (m Mode) IsDisabled(other Mode) bool {
	return ModeDisabled == other
}
func (m Mode) IsEnabled(other Mode) bool {
	return ModeEnabled == other
}

func GetMode() Mode {
	env := os.Getenv(constants.CliTerminalMode)
	return Mode(env)
}

func isTerminal() bool {
	env := GetMode()

	switch env {
	case ModeAuto, "":
		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	case ModeDisabled:
		return false
	case ModeEnabled:
		return true
	default:

		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	}
}

var _isTerminal *bool

func IsTerminal() bool {
	if _isTerminal != nil {
		return *_isTerminal
	}
	_isTerminal = new(bool)
	*_isTerminal = isTerminal()
	return *_isTerminal
}

// GetSize Tries to get terminal size from stdout file descriptor.
// on err returns -1, -1
func GetSize() (width int, height int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return -1, -1
	}
	return width, height
}
