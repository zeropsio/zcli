//go:build darwin
// +build darwin

package constants

import (
	"os"
	"path/filepath"
)

func getDataFilePathsReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliDataFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, CliDataFileName),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, CliDataFileName),
		receiverFromOsFunc(os.UserHomeDir, "zerops."+CliDataFileName),
		receiverFromOsTemp("zerops." + CliDataFileName),
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliLogFilePathEnvVar),
		receiverFromPath(filepath.Join("/usr/local/var/log/", ZeropsLogFile)),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, "zerops."+ZeropsLogFile),
		receiverFromOsTemp("zerops." + ZeropsLogFile),
	}
}

func getWgConfigFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliWgConfigPathEnvVar),
		receiverFromPath(filepath.Join("/etc/wireguard/", WgConfigFile)),
		receiverFromPath(filepath.Join("/usr/local/etc/wireguard/", WgConfigFile)),
		receiverFromPath(filepath.Join("/opt/homebrew/etc/wireguard/", WgConfigFile)),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, WgConfigFile),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, WgConfigFile),
		receiverFromOsFunc(os.UserHomeDir, WgConfigFile),
		receiverFromOsTemp("zerops." + WgConfigFile),
	}
}
