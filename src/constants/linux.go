//go:build linux
// +build linux

package constants

const (
	LogFilePath           = "/var/log/zerops/zerops.log"
	DaemonAddress         = "/run/zerops/daemon.sock"
	DaemonStorageFilePath = "/var/lib/zerops/daemon.data"
	DaemonInstallDir      = "/usr/local/"
)
