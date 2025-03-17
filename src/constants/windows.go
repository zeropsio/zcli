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
		receiverFromOsFunc(os.UserHomeDir, "zerops."+CliDataFileName),
		receiverFromOsTemp("zerops." + CliDataFileName),
	}
}

func getLogFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliLogFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", ZeropsLogFile),
		receiverFromOsFunc(os.UserHomeDir, "zerops."+ZeropsLogFile),
		receiverFromOsTemp("zerops." + ZeropsLogFile),
	}
}

func getWgConfigFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliWgConfigPathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", WgConfigFile),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", WgConfigFile),
		receiverFromOsFunc(os.UserHomeDir, WgConfigFile),
		receiverFromOsTemp("zerops." + WgConfigFile),
	}
}

func getZcliYamlFilePathsReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromOsFunc(os.UserConfigDir, "Zerops", CliZcliYamlFileName),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", CliZcliYamlFileName),
	}
}
