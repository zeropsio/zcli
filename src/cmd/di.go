package cmd

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/zerops-io/zcli/src/cliStorage"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dnsServer"
	"github.com/zerops-io/zcli/src/grpcApiClientFactory"
	"github.com/zerops-io/zcli/src/grpcDaemonServer"
	"github.com/zerops-io/zcli/src/utils/certReader"
	"github.com/zerops-io/zcli/src/utils/logger"
	"github.com/zerops-io/zcli/src/utils/tlsConfig"
	"github.com/zerops-io/zcli/src/vpn"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func createCertReader(token string) (*certReader.Handler, error) {
	return certReader.New(
		certReader.Config{
			Token: token,
		},
	)
}

func createCliStorage() (*cliStorage.Handler, error) {
	return cliStorage.New(
		cliStorage.Config{
			FilePath: constants.CliStorageFile,
		},
	)
}

func createTlsConfig(certReader *certReader.Handler) (*tls.Config, error) {
	return tlsConfig.CreateTlsConfig(
		certReader,
	)
}

func createApiGrpcClient(ctx context.Context, tlsConfig *tls.Config) (_ zeropsApiProtocol.ZeropsApiProtocolClient, closeFunc func(), _ error) {

	connection, err := grpc.DialContext(ctx, params.GetPersistentString("grpcApiAddress"), grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return nil, nil, err
	}

	closeFunc = func() { _ = connection.Close() }

	return zeropsApiProtocol.NewZeropsApiProtocolClient(connection), closeFunc, nil

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

func createVpn(storage *daemonStorage.Handler, dnsServer *dnsServer.Handler, logger *logger.Handler) *vpn.Handler {
	return vpn.New(
		vpn.Config{
			VpnCheckInterval:   time.Second * 3,
			VpnCheckRetryCount: 3,
			VpnCheckTimeout:    time.Second * 3,
		},
		logger,
		grpcApiClientFactory.New(),
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
