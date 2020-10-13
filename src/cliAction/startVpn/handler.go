package startVpn

import (
	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/grpcDaemonClientFactory"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

type Config struct {
	GrpcApiAddress string
	VpnAddress     string
}

type RunConfig struct {
	ProjectName string
	Token       string
	Mtu         uint32
}

type Handler struct {
	config                    Config
	apiGrpcClient             zeropsApiProtocol.ZeropsApiProtocolClient
	zeropsDaemonClientFactory *grpcDaemonClientFactory.Handler
	daemonInstaller           *daemonInstaller.Handler
}

func New(
	config Config,
	apiGrpcClient zeropsApiProtocol.ZeropsApiProtocolClient,
	zeropsDaemonClientFactory *grpcDaemonClientFactory.Handler,
	daemonInstaller *daemonInstaller.Handler,
) *Handler {
	return &Handler{
		config:                    config,
		apiGrpcClient:             apiGrpcClient,
		zeropsDaemonClientFactory: zeropsDaemonClientFactory,
		daemonInstaller:           daemonInstaller,
	}
}
