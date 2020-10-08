package grpcApiClientFactory

import (
	"context"

	"github.com/zerops-io/zcli/src/utils/certReader"
	"github.com/zerops-io/zcli/src/utils/tlsConfig"

	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) CreateClient(ctx context.Context, grpcApiAddress string, token string) (_ zeropsApiProtocol.ZeropsApiProtocolClient, closeFunc func(), err error) {

	certReader, err := certReader.New(
		certReader.Config{
			Token: token,
		},
	)
	if err != nil {
		return
	}

	tlsConfig, err := tlsConfig.CreateTlsConfig(certReader)
	if err != nil {
		return
	}

	connection, err := grpc.DialContext(
		ctx,
		grpcApiAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		return
	}

	closeFunc = func() { _ = connection.Close() }

	return zeropsApiProtocol.NewZeropsApiProtocolClient(connection), closeFunc, nil

}
