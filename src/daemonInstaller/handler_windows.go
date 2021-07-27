// +build windows

package daemonInstaller

import (
	"errors"
)

type windowsRecord struct {
	name         string
	description  string
	dependencies []string
}

func newDaemon(name, description string, dependencies []string) (daemon, error) {
	return &windowsRecord{
		name:         name,
		description:  description,
		dependencies: dependencies,
	}, nil
}

func (daemon *windowsRecord) Install() error {
	return errors.New("windows is not supported")
}

func (daemon *windowsRecord) Remove() error {
	return errors.New("windows is not supported")
}

func (daemon *windowsRecord) IsInstalled() bool {
	return false
}
