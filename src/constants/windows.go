//go:build windows
// +build windows

package constants

import (
	"os"

	"github.com/pkg/errors"
)

// this line is here to make linter happy
var _ = zeropsDir

func getDataFilePaths() []pathReceiver {
	return []pathReceiver{
		func() (string, error) {
			env := os.Getenv(cliDataFilePathEnvVar)
			if env != "" {
				return env, nil
			}
			return "", errors.New("env is empty")
		},
		receiverWithPath(os.UserConfigDir, "Zerops", cliDataFileName),
		receiverWithPath(os.UserHomeDir, "Zerops", cliDataFileName),
	}
}

func getLogFilePath() []pathReceiver {
	return []pathReceiver{
		func() (string, error) {
			env := os.Getenv(cliLogFilePathEnvVar)
			if env != "" {
				return env, nil
			}
			return "", errors.New("env is empty")
		},
		receiverWithPath(os.UserConfigDir, "Zerops", zeropsLogFile),
		receiverWithPath(os.UserHomeDir, "Zerops", zeropsLogFile),
	}
}
