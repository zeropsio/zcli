//go:build windows
// +build windows

package wgquick

import (
	"os"
	"os/exec"
)

func New() Configurator {
	return Configurator{
		configPath:  os.TempDir(),
		upCommand:   exec.Command("wireguard", "/installtunnelservice"),
		downCommand: exec.Command("wireguard", "/uninstalltunnelservice"),
	}
}
