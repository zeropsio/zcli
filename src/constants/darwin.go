//go:build darwin
// +build darwin

package constants

import (
	"os"
	"path"

	"github.com/pkg/errors"
)

func getDataFilePaths() []pathReceiver {
	return []pathReceiver{
		func() (string, error) {
			env := os.Getenv(cliDataFilePathEnvVar)
			if env != "" {
				return env, nil
			}
			return "", errors.New("env is empty")
		},
		receiverWithPath(os.UserConfigDir, zeropsDir, cliDataFileName),
		receiverWithPath(os.UserHomeDir, zeropsDir, cliDataFileName),
	}
}

func getLogFilePath() []pathReceiver {
	return []pathReceiver{
		func() (string, error) {
			return path.Join("/usr/local/var/log/", zeropsLogFile), nil
		},
		receiverWithPath(os.UserConfigDir, zeropsDir, zeropsLogFile),
		receiverWithPath(os.UserHomeDir, zeropsDir, zeropsLogFile),
	}
}
