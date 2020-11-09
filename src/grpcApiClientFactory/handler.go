package grpcApiClientFactory

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/credentials/oauth"

	"google.golang.org/grpc/credentials"

	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"google.golang.org/grpc"
)

type Config struct {
	CaCertificateUrl string
}

type Handler struct {
	config Config
}

func New(
	config Config,
) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) CreateClient(ctx context.Context, grpcApiAddress string, token string) (_ zeropsApiProtocol.ZeropsApiProtocolClient, closeFunc func(), err error) {

	tlsCreds, err := h.createTLSCredentials()
	if err != nil {
		return nil, nil, err
	}
	connection, err := grpc.DialContext(
		ctx,
		grpcApiAddress,
		grpc.WithPerRPCCredentials(h.createBearerCredentials(token)),
		grpc.WithTransportCredentials(tlsCreds),
	)
	if err != nil {
		return
	}

	closeFunc = func() { _ = connection.Close() }

	return zeropsApiProtocol.NewZeropsApiProtocolClient(connection), closeFunc, nil

}

func (h *Handler) createBearerCredentials(token string) credentials.PerRPCCredentials {
	return oauth.NewOauthAccess(&oauth2.Token{AccessToken: token, TokenType: "Bearer"})
}

func (h *Handler) createTLSCredentials() (credentials.TransportCredentials, error) {

	resp, err := http.Get(h.config.CaCertificateUrl)
	if err != nil {
		return nil, fmt.Errorf("get caCertificate => %s", err.Error())
	}
	defer resp.Body.Close()
	caCertBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read caCertificate response => %s", err.Error())
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertBytes) {
		return nil, fmt.Errorf("failed to add server CA certificate")
	}
	config := &tls.Config{
		RootCAs:    certPool,
		ServerName: "zbusinessapi-runtime-1-2",
	}
	return credentials.NewTLS(config), nil
}
