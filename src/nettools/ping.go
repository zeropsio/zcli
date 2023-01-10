package nettools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func HasIPv6PingCommand() bool {
	if hasPing6Command() {
		return true
	}
	if _, err := exec.LookPath("ping"); err == nil {
		return true
	}
	return runtime.GOOS == "windows"
}

func hasPing6Command() bool {
	_, err := exec.LookPath("ping6")
	return err == nil
}

func Ping(ctx context.Context, address string) error {
	cmd := []string{"ping6", "-c", "1", address}
	if !hasPing6Command() {
		cmd = []string{"ping", "-6", "-c", "1", address}
	}
	if runtime.GOOS == "windows" {
		cmd = []string{"ping", "/n", "1", address}
	}

	if out, err := exec.CommandContext(ctx, cmd[0], cmd[1:]...).Output(); err != nil {
		return pingError(err, cmd, out)
	}
	return nil
}

type PingError struct {
	err    error
	cmd    []string
	output string
}

func (p PingError) Err() error {
	return p.err
}

func (p PingError) Output() string {
	return p.output
}

func (p PingError) Cmd() string {
	return strings.Join(p.cmd, " ")
}

func (p PingError) Error() string {
	return fmt.Sprintf("ping => err: %s, exec: %s, output: %s", p.err, p.Cmd(), p.output)
}

func pingError(
	err error,
	cmd []string,
	output []byte,
) error {
	return PingError{
		err:    err,
		cmd:    cmd,
		output: string(output),
	}
}
