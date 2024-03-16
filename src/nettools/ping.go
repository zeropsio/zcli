package nettools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func Ping(ctx context.Context, address string) error {
	cmd := []string{"-c", "1", address}

	if runtime.GOOS == "windows" {
		cmd = []string{"/n", "1", address}
	}

	if out, err := exec.CommandContext(ctx, "ping", cmd...).Output(); err != nil {
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
