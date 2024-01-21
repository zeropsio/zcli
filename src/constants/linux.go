//go:build linux
// +build linux

package constants

func getDataFilePaths() []pathReceiver {
	return []pathReceiver{
		receiverWithPath(os.UserConfigDir, zeropsDir, cliDataFileName),
		receiverWithPath(os.UserHomeDir, zeropsDir, cliDataFileName),
	}
}

func getLogFilePath() []pathReceiver {
	return []pathReceiver{
		func() (string, error) {
			return path.Join("/var/log/", zeropsDir, zeropsLogFile), nil
		},
		receiverWithPath(os.UserConfigDir, zeropsDir, zeropsLogFile),
		receiverWithPath(os.UserHomeDir, zeropsDir, zeropsLogFile),
	}
}
