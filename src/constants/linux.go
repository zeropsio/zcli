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
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliLogFilePathEnvVar),
		receiverFromPath(path.Join("/var/log/", ZeropsLogFile)),
		receiverFromOsFunc(os.UserConfigDir, ZeropsDir, ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, ZeropsDir, ZeropsLogFile),
	}
}
