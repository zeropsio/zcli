package nettools

import (
	"context"
	"os/exec"
	"runtime"
)

func HasIPv6PingCommand() bool {
	_, err := exec.LookPath("ping6")
	return err == nil || runtime.GOOS == "windows"
}

func Ping(ctx context.Context, address string) error {
	pingCommand := exec.CommandContext(ctx, "ping6", "-c", "1", address)
	if runtime.GOOS == "windows" {
		pingCommand = exec.CommandContext(ctx, "ping", "/n", "1", address)
	}

	return pingCommand.Run()
}
