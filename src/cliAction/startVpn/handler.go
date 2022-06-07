package startVpn

import (
	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
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
	sdkConfig       sdkConfig.Config
}

func New(
	config Config,
	apiGrpcClient business.ZeropsApiProtocolClient,
	daemonInstaller *daemonInstaller.Handler,
	sdkConfig sdkConfig.Config,
) *Handler {
	return &Handler{
		config:          config,
		apiGrpcClient:   apiGrpcClient,
		daemonInstaller: daemonInstaller,
		sdkConfig:       sdkConfig,
	}
}
