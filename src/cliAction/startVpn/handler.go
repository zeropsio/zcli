package startVpn

import (
	"github.com/zeropsio/zcli/src/daemonInstaller"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
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
	PreferredPortMin uint32
	PreferredPortMax uint32
}

type Handler struct {
	config          Config
	apiGrpcClient   zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient
	daemonInstaller *daemonInstaller.Handler
	sdkConfig       sdkConfig.Config
}

func New(
	config Config,
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient,
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
