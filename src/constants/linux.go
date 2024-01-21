//go:build linux
// +build linux

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
			env := os.Getenv(cliLogFilePathEnvVar)
			if env != "" {
				return env, nil
			}
			return "", errors.New("env is empty")
		},
		func() (string, error) {
			return path.Join("/var/log/", zeropsDir, zeropsLogFile), nil
		},
		receiverWithPath(os.UserConfigDir, zeropsDir, zeropsLogFile),
		receiverWithPath(os.UserHomeDir, zeropsDir, zeropsLogFile),
	}
}
