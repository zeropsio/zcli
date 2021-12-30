package cmdRunner

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
)

var IpAlreadySetErr = errors.New("RTNETLINK answers: File exists")
var CannotFindDeviceErr = errors.New(`Cannot find device "wg0"`)
var OperationNotPermitted = errors.New(`Operation not permitted`)

type ExecErrInterface interface {
	error
	ExitCode() int
}

type execErr struct {
	cmd      *exec.Cmd
	prev     error
	exitCode int
}

func (e execErr) Error() string {
	return e.cmd.String() + ": " + e.prev.Error()
}

func (e execErr) ExitCode() int {
	return e.exitCode
}

func (e execErr) Unwrap() error {
	return e.prev
}

func (e execErr) Is(target error) bool {
	return errors.Is(e.prev, target)
}

func Run(cmd *exec.Cmd) ([]byte, ExecErrInterface) {
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = errOutput
	cmd.Env = append(os.Environ(), cmd.Env...)

	if err := cmd.Run(); err != nil {
		exitCode := 0
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		execError := &execErr{
			cmd:      cmd,
			exitCode: exitCode,
			prev:     err,
		}

		if errOutput.Len() == 0 {
			return nil, execError
		}

		errOutputString := string(errOutput.Bytes()[0 : errOutput.Len()-1])

		if strings.Contains(errOutputString, OperationNotPermitted.Error()) {
			execError.prev = OperationNotPermitted
			return nil, execError
		}

		for _, e := range []error{IpAlreadySetErr, CannotFindDeviceErr} {
			if errOutputString == e.Error() {
				execError.prev = e
				return nil, execError
			}
		}

		execError.prev = errors.New(errOutputString)
		return nil, execError
	}

	return output.Bytes(), nil
}
