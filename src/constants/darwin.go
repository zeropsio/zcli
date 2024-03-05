//go:build darwin
// +build darwin

package constants

import (
	"os"
	"path"
)

func getDataFilePathsReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliDataFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, CliDataFileName),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, CliDataFileName),
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliLogFilePathEnvVar),
		receiverFromPath(path.Join("/usr/local/var/log/", ZeropsLogFile)),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, ZeropsLogFile),
	}
}

func getWgConfigFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliWgConfigPathEnvVar),
		receiverFromPath(path.Join("/etc/wireguard/", WgConfigFile)),
		receiverFromPath(path.Join("/usr/local/etc/wireguard/", WgConfigFile)),
		receiverFromPath(path.Join("/opt/homebrew/etc/wireguard/", WgConfigFile)),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, WgConfigFile),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, WgConfigFile),
	}
}
