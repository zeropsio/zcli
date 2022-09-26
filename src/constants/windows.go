//go:build windows
// +build windows

package constants

import (
	"os"
	"path/filepath"
)

var (
	LogFilePath,
	DaemonStorageFilePath,
	DaemonAddress,
	DaemonInstallDir string
)

const WireguardPath = `C:\Program Files\wireguard`

func init() {
	appData, _ := os.UserConfigDir()
	zeropsFolder := filepath.Join(appData, "Zerops")

	LogFilePath = filepath.Join(zeropsFolder, "zerops.log")
	DaemonAddress = ":45677"
	DaemonStorageFilePath = filepath.Join(zeropsFolder, "daemon.data")
	DaemonInstallDir = zeropsFolder
}
