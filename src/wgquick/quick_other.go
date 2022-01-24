//go:build !windows
// +build !windows

package wgquick

import "os/exec"

func New() Configurator {
	return Configurator{
		configPath:  "/etc/wireguard",
		downCommand: exec.Command("wg-quick", "down"),
		upCommand:   exec.Command("wg-quick", "up"),
	}
}
