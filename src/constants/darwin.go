// +build darwin

package constants

import (
	"os"
	"path"
)

const (
	LogFilePath           = "/usr/local/var/log/zerops.log"
	SocketFilePath        = "/usr/local/var/zerops/daemon.sock"
	DaemonStorageFilePath = "/usr/local/var/zerops/daemon.data"
)

var (
	CliStorageFile string
)

func init() {
	CliStorageFile = path.Join(os.Getenv("HOME"), "/.config/zerops/cli.data")
}
