//go:build linux
// +build linux

package constants

import (
	"os"
	"path"
)

func getDataFilePathsReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(cliDataFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, zeropsDir, cliDataFileName),
		receiverFromOsFunc(os.UserHomeDir, zeropsDir, cliDataFileName),
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(cliLogFilePathEnvVar),
		receiverFromPath(path.Join("/var/log/", zeropsLogFile)),
		receiverFromOsFunc(os.UserConfigDir, zeropsDir, zeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, zeropsDir, zeropsLogFile),
	}
}
