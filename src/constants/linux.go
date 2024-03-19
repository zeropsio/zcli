//go:build linux
// +build linux

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
		receiverFromOsFunc(os.UserHomeDir, "zerops."+CliDataFileName),
		receiverFromOsTemp("zerops." + CliDataFileName),
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliLogFilePathEnvVar),
		receiverFromPath(path.Join("/var/log/", ZeropsLogFile)),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, "zerops."+ZeropsLogFile),
		receiverFromOsTemp("zerops." + ZeropsLogFile),
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
		receiverFromOsFunc(os.UserHomeDir, WgConfigFile),
		receiverFromOsTemp("zerops." + WgConfigFile),
	}
}
