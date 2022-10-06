package constants

import (
	"os"
	"path/filepath"
)

func CliLoginData() (string, error) {
	return cliStorageFilepath("cli.data")
}

func CliRegionData() (string, error) {
	return cliStorageFilepath("region.data")
}

func CliStorageFilepath() (string, error) {
	return cliStorageFilepath("")
}

func cliStorageFilepath(filename string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "zerops", filename), nil
}
