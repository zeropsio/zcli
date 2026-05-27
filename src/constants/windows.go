//go:build windows

package constants

import (
	"os"
)

// this is here to make linter happy
var _ = ZeropsDir
var _ = receiverFromPath

func getProfileFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		receiverFromEnv(CliDataFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", CliProfileFileName),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", CliProfileFileName),
		receiverFromOsFunc(os.UserHomeDir, "zerops."+CliProfileFileName),
		receiverFromOsTemp("zerops." + CliProfileFileName),
	}
}

func getLegacyDataFilePathReceivers() []pathReceiver {
	return []pathReceiver{
		existingReceiverFromOsFunc(os.UserConfigDir, "Zerops", CliDataFileNameOld),
		existingReceiverFromOsFunc(os.UserHomeDir, "Zerops", CliDataFileNameOld),
		existingReceiverFromOsFunc(os.UserHomeDir, "zerops."+CliDataFileNameOld),
		existingReceiverFromOsTemp("zerops." + CliDataFileNameOld),
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
		receiverFromEnv(CliZcliYamlFilePathEnvVar),
		receiverFromOsFunc(os.UserConfigDir, "Zerops", CliZcliYamlFileName),
		receiverFromOsFunc(os.UserHomeDir, "Zerops", CliZcliYamlFileName),
	}
}
