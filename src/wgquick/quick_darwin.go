//go:build darwin
// +build darwin

package wgquick

import (
	"os/exec"

	"github.com/zerops-io/zcli/src/constants"
)

func New() Configurator {
	return Configurator{
		configPath:    constants.WireguardConfigDir,
		downCommand:   exec.Command("wg-quick", "down"),
		upCommand:     exec.Command("wg-quick", "up"),
		additionalDns: []string{"8.8.8.8"},
	}
}
