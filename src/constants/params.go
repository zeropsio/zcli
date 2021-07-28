package constants

import (
	"os"
	"path"
)

const (
	PersistentParamCaCertificateUrl = "caCertificateUrl"
	PersistentParamRestApiAddress   = "restApiAddress"
	PersistentParamGrpcApiAddress   = "grpcApiAddress"
	PersistentParamVpnApiAddress    = "vpnApiAddress"
)

func CliStorageFile() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(configDir, ".config", "zerops", "cli.data"), nil
}
