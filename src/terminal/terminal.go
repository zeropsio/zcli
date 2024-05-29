package terminal

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/zeropsio/zcli/src/constants"
)

type terminalMode string

const (
	ModeAuto     terminalMode = "auto"
	ModeDisabled terminalMode = "disabled"
	ModeEnabled  terminalMode = "enabled"
)

func isTerminal() bool {
	env := os.Getenv(constants.CliTerminalMode)

	switch terminalMode(env) {
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
