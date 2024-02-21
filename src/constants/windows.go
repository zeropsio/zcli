//go:build windows
// +build windows

package constants

import (
	"os"
)

// this is here to make linter happy
var _ = zeropsDir
var _ = receiverFromPath

func getDataFilePathsReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(cliDataFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", cliDataFileName),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", cliDataFileName),
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(cliLogFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", zeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", zeropsLogFile),
	}
}
