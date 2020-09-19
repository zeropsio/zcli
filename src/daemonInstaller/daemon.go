package daemonInstaller

import "errors"

type daemon interface {
	Install() error
	Remove() error
	IsInstalled() bool
}

var (
	ErrAlreadyInstalled = errors.New("Service has already been installed")
	ErrNotInstalled     = errors.New("Service is not installed")
)
