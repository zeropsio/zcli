package sudoers

import (
	"errors"
	"os/exec"

	"github.com/zerops-io/zcli/src/helpers/cmdRunner"
)

var IpAlreadySetErr = cmdRunner.IpAlreadySetErr
var CannotFindDeviceErr = cmdRunner.CannotFindDeviceErr
var OperationNotPermitted = cmdRunner.OperationNotPermitted

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

	sudoCmd := exec.Command("sudo", cmd.Args...)
	sudoCmd.Env = cmd.Env
	sudoCmd.Stdin = cmd.Stdin
	sudoCmd.Stderr = cmd.Stderr
	sudoCmd.Stdout = cmd.Stdout
	sudoCmd.Dir = cmd.Dir

	output, err := cmdRunner.Run(sudoCmd)
	if err != nil {
		if errors.Is(err, OperationNotPermitted) {

			newCmd := exec.Command("sudo", cmd.Args...)

			output, err = cmdRunner.Run(newCmd)
		}
	}

	return output, err

}
