//go:build linux
// +build linux

package daemonInstaller

import (
	"errors"
	"os"
)

func newDaemon(name, description string, dependencies []string) (daemon, error) {

	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return &systemDRecord{
			name:         name,
			description:  description,
			dependencies: dependencies,
		}, nil
	}
	return nil, errors.New("systemd is not installed")
}
