package sudoers

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

var IpAlreadySetErr = errors.New("RTNETLINK answers: File exists")
var CannotFindDeviceErr = errors.New(`Cannot find device "wg0"`)
var OperationNotPermitted = errors.New(`Operation not permitted`)

type Config struct {
}

type Handler struct {
	config Config
}

func New(config Config) *Handler {
	return &Handler{
		config: config,
	}
}

// command with installation if operation is not permitted
func (h *Handler) RunCommand(cmd *exec.Cmd) ([]byte, error) {

	output, err := runCommand(cmd)
	if err != nil {
		if errors.Is(err, OperationNotPermitted) {

			newCmd := exec.Command("sudo", cmd.Args...)

			output, err = runCommand(newCmd)
		}
	}

	return output, err

}

func runCommand(cmd *exec.Cmd) ([]byte, error) {
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = errOutput

	if err := cmd.Run(); err != nil {
		if errOutput.Len() > 0 {
			errOutputString := string(errOutput.Bytes()[0 : errOutput.Len()-1])

			if strings.Contains(errOutputString, OperationNotPermitted.Error()) {
				return nil, OperationNotPermitted
			}

			for _, e := range []error{IpAlreadySetErr, CannotFindDeviceErr} {
				if errOutputString == e.Error() {
					return nil, e
				}
			}
		}
		return nil, errors.New(errOutput.String())
	}

	return output.Bytes(), nil
}
