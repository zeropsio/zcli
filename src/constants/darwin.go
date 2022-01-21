//go:build darwin
// +build darwin

package constants

const (
	LogFilePath           = "/usr/local/var/log/zerops.log"
	DaemonAddress         = "/usr/local/var/zerops/daemon.sock"
	DaemonStorageFilePath = "/usr/local/var/zerops/daemon.data"
	DaemonInstallDir      = "/usr/local/"
)
