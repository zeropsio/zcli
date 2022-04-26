package cmd

import (
	"github.com/zerops-io/zcli/src/prolongVpn"
	"time"

	"github.com/zerops-io/zcli/src/cliStorage"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dnsServer"
	"github.com/zerops-io/zcli/src/region"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/logger"
	"github.com/zerops-io/zcli/src/utils/storage"
	"github.com/zerops-io/zcli/src/vpn"
)

func getToken(storage *cliStorage.Handler) string {
	token := BuiltinToken
	if storage.Data().Token != "" {
		token = storage.Data().Token
	}

	return token
}

func createLogger() (*logger.Handler, error) {
	return logger.New(logger.Config{
		FilePath: constants.LogFilePath,
	})
}

func createCliStorage() (*cliStorage.Handler, error) {
	filePath, err := constants.CliLoginData()
	if err != nil {
		return nil, err
	}
	s, err := storage.New[cliStorage.Data](
		storage.Config{
			FilePath: filePath,
		},
	)
	return &cliStorage.Handler{Handler: *s}, err
}

func createDnsServer() *dnsServer.Handler {
	return dnsServer.New()
}

func createDaemonStorage() (*daemonStorage.Handler, error) {
	s, err := storage.New[daemonStorage.Data](
		storage.Config{
			FilePath: constants.DaemonStorageFilePath,
		},
	)
	return &daemonStorage.Handler{Handler: *s}, err
}

func createRegionRetriever() (*region.Handler, error) {
	filepath, err := constants.CliRegionData()
	if err != nil {
		return nil, err
	}
	s, err := storage.New[region.Data](
		storage.Config{FilePath: filepath},
	)
	return region.New(httpClient.New(httpClient.Config{HttpTimeout: time.Second * 5}), s), err
}

func createVpn(
	storage *daemonStorage.Handler,
	dnsServer *dnsServer.Handler,
	logger *logger.Handler,
) *vpn.Handler {
	return vpn.New(
		vpn.Config{
			VpnCheckInterval:   time.Second * 3,
			VpnCheckRetryCount: 3,
			VpnCheckTimeout:    time.Second * 3,
		},
		logger,
		storage,
		dnsServer,
	)
}

func createVpnProlong(
	storage *daemonStorage.Handler,
	logger *logger.Handler,
) *prolongVpn.Handler {
	return prolongVpn.New(storage, logger)
}
