package vpn

import (
	"context"
	"errors"
	"os/exec"

	"github.com/zerops-io/zcli/src/utils/logger"
)

type command struct {
	command      string
	args         []string
	errorMessage string
}

func makeCommand(cmd string, errorMessage string, args ...string) command {
	return command{
		command:      cmd,
		args:         args,
		errorMessage: errorMessage,
	}
}

func runCommands(
	ctx context.Context,
	logger logger.Logger,
	cmds ...command,
) error {
	for _, cmd := range cmds {
		if output, err := exec.CommandContext(ctx, cmd.command, cmd.args...).CombinedOutput(); err != nil {
			logger.Debug(cmd.command, cmd.args)
			logger.Debug(string(output))
			logger.Error(err)
			return errors.New(cmd.errorMessage)

		}
	}
	return nil
}
