//go:build windows
// +build windows

package constants

import (
	"os"
)

// this is here to make linter happy
var _ = ZeropsDir
var _ = receiverFromPath

func getDataFilePathsReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliDataFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", CliDataFileName),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", CliDataFileName),
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliLogFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", ZeropsLogFile),
	}
}
