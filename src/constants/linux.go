// +build linux

package constants

import (
	"os"
	"path"
)

const (
	LogFilePath           = "/var/log/zerops/zerops.log"
	SocketFilePath        = "/run/zerops/daemon.sock"
	DaemonStorageFilePath = "/var/lib/zerops/daemon.data"
	DaemonInstallDir      = "/usr/local/"
)

func CliStorageFile() string {
	return path.Join(os.Getenv("HOME"), "/.config/zerops/cli.data")
}
