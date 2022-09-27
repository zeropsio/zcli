package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/daemonStorage"
	"github.com/zeropsio/zcli/src/dnsServer"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/prolongVpn"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/logger"
	"github.com/zeropsio/zcli/src/utils/storage"
	"github.com/zeropsio/zcli/src/vpn"
)

func getToken(storage *cliStorage.Handler) (string, error) {
	token := BuiltinToken
	if storage.Data().Token != "" {
		token = storage.Data().Token
	}
	if token == "" {
		return token, errors.New(i18n.UnauthenticatedUser)
	}
	return token, nil
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
	return &cliStorage.Handler{Handler: s}, err
}

func createDnsServer(logger logger.Logger) *dnsServer.Handler {
	return dnsServer.New(logger)
}

func createDaemonStorage() (*daemonStorage.Handler, error) {
	s, err := storage.New[daemonStorage.Data](
		storage.Config{
			FilePath: constants.DaemonStorageFilePath,
		},
	)
	return &daemonStorage.Handler{Handler: s}, err
}

func createRegionRetriever(ctx context.Context) (*region.Handler, error) {
	filepath, err := constants.CliRegionData()
	if err != nil {
		return nil, err
	}
	s, err := storage.New[region.Data](
		storage.Config{FilePath: filepath},
	)
	return region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}), s), err
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
			VpnCheckTimeout:    time.Second * 60,
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
