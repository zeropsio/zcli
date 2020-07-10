package cmd

import (
	"context"
	"crypto/tls"
	"os"
	"path"

	"github.com/zerops-io/zcli/src/service/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/zerops-io/zcli/src/service/sudoers"

	"github.com/zerops-io/zcli/src/service/certReader"
	"github.com/zerops-io/zcli/src/service/tlsConfig"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func createCertReader() (*certReader.Handler, error) {
	return certReader.New(
		certReader.Config{
			Token: params.GetString("token"),
		},
	)
}

func createSudoers() *sudoers.Handler {
	return sudoers.New(
		sudoers.Config{},
	)
}

func createStorage() (*storage.Handler, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return storage.New(
		storage.Config{
			FilePath: path.Join(currentDir, "zcli.data"),
		},
	)
}

func createTlsConfig(certReader *certReader.Handler) (*tls.Config, error) {
	return tlsConfig.CreateTlsConfig(
		certReader,
	)
}

func createApiGrpcClient(ctx context.Context, tlsConfig *tls.Config) (_ zeropsApiProtocol.ZeropsApiProtocolClient, closeFunc func(), _ error) {

	connection, err := grpc.DialContext(ctx, params.GetString("grpcApiAddress"), grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return nil, nil, err
	}

	closeFunc = func() { _ = connection.Close() }

	return zeropsApiProtocol.NewZeropsApiProtocolClient(connection), closeFunc, nil

}
