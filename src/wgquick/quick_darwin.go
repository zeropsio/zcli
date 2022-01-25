//go:build darwin
// +build darwin

package wgquick

import "os/exec"

func New() Configurator {
	return Configurator{
		configPath:    "/etc/wireguard",
		downCommand:   exec.Command("wg-quick", "down"),
		upCommand:     exec.Command("wg-quick", "up"),
		additionalDns: []string{"8.8.8.8"},
	}
}
