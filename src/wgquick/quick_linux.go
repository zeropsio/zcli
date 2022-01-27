//go:build linux
// +build linux

package wgquick

import (
	"os/exec"

	"github.com/zerops-io/zcli/src/constants"
)

func New() Configurator {
	return Configurator{
		configPath:  constants.WireguardConfigDir,
		downCommand: exec.Command("wg-quick", "down"),
		upCommand:   exec.Command("wg-quick", "up"),
	}
}
