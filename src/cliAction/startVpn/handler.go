package startVpn

import (
	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/proto/business"
)

type Config struct {
	GrpcApiAddress string
	VpnAddress     string
}

type RunConfig struct {
	ProjectNameOrId  string
	Token            string
	Mtu              uint32
	CaCertificateUrl string
}

type Handler struct {
	config          Config
	apiGrpcClient   business.ZeropsApiProtocolClient
	daemonInstaller *daemonInstaller.Handler
}

func New(
	config Config,
	apiGrpcClient business.ZeropsApiProtocolClient,
	daemonInstaller *daemonInstaller.Handler,
) *Handler {
	return &Handler{
		config:          config,
		apiGrpcClient:   apiGrpcClient,
		daemonInstaller: daemonInstaller,
	}
}
