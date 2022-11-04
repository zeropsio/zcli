package vpn

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

type command struct {
	command      string
	args         []string
	stdin        string
	errorMessage string
}

func commandWithArgs(args ...string) func(*command) {
	return func(c *command) {
		c.args = append(c.args, args...)
	}
}
func commandWithErrorMessage(errorMessage string) func(*command) {
	return func(c *command) {
		c.errorMessage = errorMessage
	}
}

func commandWithStdin(stdin string) func(*command) {
	return func(c *command) {
		c.stdin = stdin
	}
}

func makeCommand(cmd string, options ...func(*command)) command {
	c := &command{
		command: cmd,
	}
	for _, o := range options {
		o(c)
	}
	return *c
}

func (h *Handler) runCommands(
	ctx context.Context,
	commands ...command,
) error {
	for _, cmd := range commands {
		_, err := h.runCommand(ctx, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) runCommand(
	ctx context.Context,
	cmd command,
) ([]byte, error) {
	c := exec.CommandContext(ctx, cmd.command, cmd.args...)
	if cmd.stdin != "" {
		c.Stdin = strings.NewReader(cmd.stdin)
	}
	output, err := c.Output()
	if err != nil {
		h.logger.Debug(cmd.command, cmd.args)
		h.logger.Debug(string(output))
		h.logger.Error(err)
		return nil, errors.New(cmd.errorMessage)
	} else {
		h.logger.Debug(cmd.command, cmd.args)
		h.logger.Debug(string(output))
	}
	return output, err
}
