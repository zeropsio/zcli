package cmdRunner

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
)

var ErrIpAlreadySet = errors.New("RTNETLINK answers: File exists")
var ErrCannotFindDevice = errors.New(`Cannot find device "wg0"`)
var ErrOperationNotPermitted = errors.New(`Operation not permitted`)

type execError struct {
	cmd      *ExecCmd
	prev     error
	exitCode int
}

func (e execError) Error() string {
	return e.cmd.String() + ": " + e.prev.Error()
}

func (e execError) ExitCode() int {
	return e.exitCode
}

func (e execError) Unwrap() error {
	return e.prev
}

func (e execError) Is(target error) bool {
	return errors.Is(e.prev, target)
}

func Run(cmd *ExecCmd) ([]byte, error) {
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = errOutput
	cmd.Env = append(os.Environ(), cmd.Env...)

	if err := cmd.execBefore(); err != nil {
		return nil, err
	}

	if err := cmd.Run(); err != nil {
		exitCode := 0
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}

		execError := &execError{
			cmd:      cmd,
			exitCode: exitCode,
			prev:     err,
		}

		if errOutput.Len() == 0 {
			return nil, execError
		}

		errOutputString := string(errOutput.Bytes()[0 : errOutput.Len()-1])

		if strings.Contains(errOutputString, ErrOperationNotPermitted.Error()) {
			execError.prev = ErrOperationNotPermitted
			return nil, execError
		}

		for _, e := range []error{ErrIpAlreadySet, ErrCannotFindDevice} {
			if errOutputString == e.Error() {
				execError.prev = e
				return nil, execError
			}
		}

		execError.prev = errors.New(errOutputString)
		return nil, execError
	}

	if err := cmd.execAfter(); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}
