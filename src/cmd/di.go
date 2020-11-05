package cmd

import (
	"time"

	"github.com/zerops-io/zcli/src/cliStorage"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dnsServer"
	"github.com/zerops-io/zcli/src/grpcDaemonServer"
	"github.com/zerops-io/zcli/src/utils/logger"
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
	return cliStorage.New(
		cliStorage.Config{
			FilePath: constants.CliStorageFile(),
		},
	)
}

func createDnsServer() *dnsServer.Handler {
	return dnsServer.New()
}

func createDaemonStorage() (*daemonStorage.Handler, error) {
	return daemonStorage.New(
		daemonStorage.Config{
			FilePath: constants.DaemonStorageFilePath,
		},
	)
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

func createDaemonGrpcServer(vpn *vpn.Handler) *grpcDaemonServer.Handler {
	return grpcDaemonServer.New(grpcDaemonServer.Config{
		Socket: constants.SocketFilePath,
	},
		vpn,
	)
}
