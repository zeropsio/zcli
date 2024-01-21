//go:build windows
// +build windows

package constants

import (
	"os"
	"path"
)

func getDataFilePaths() []pathReceiver {
	return []pathReceiver{
		receiverWithPath(os.UserConfigDir, "Zerops", cliDataFileName),
		receiverWithPath(os.UserHomeDir, "Zerops", cliDataFileName),
	}
}

func getLogFilePath() []pathReceiver {
	return []pathReceiver{
		func() (string, error) {
			return path.Join("/usr/local/var/log/", zeropsLogFile), nil
		},
		receiverWithPath(os.UserConfigDir, "Zerops", zeropsLogFile),
		receiverWithPath(os.UserHomeDir, "Zerops", zeropsLogFile),
	}
}
