package stopVpn

import (
	"context"
	"fmt"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/daemon"

	"github.com/zerops-io/zcli/src/i18n"
)

type Config struct {
}

type RunConfig struct {
}

type Handler struct {
	config             Config
	zeropsDaemonClient daemon.ZeropsDaemonProtocolClient
}

func New(
	config Config,
	zeropsDaemonClient daemon.ZeropsDaemonProtocolClient,
) *Handler {
	return &Handler{
		config:             config,
		zeropsDaemonClient: zeropsDaemonClient,
	}
}

func (h *Handler) Run(ctx context.Context, _ RunConfig) error {

	response, err := h.zeropsDaemonClient.StopVpn(ctx, &daemon.StopVpnRequest{})
	daemonInstalled, err := proto.DaemonError(err)
	if err != nil {
		return err
	}

	if !daemonInstalled {
		fmt.Println(i18n.VpnStopDaemonIsUnavailable)
		return nil
	}

	if response.GetActiveBefore() {
		fmt.Println(i18n.VpnStopSuccess)
		if response.GetVpnStatus().GetAdditionalInfo() != "" {
			fmt.Println(i18n.VpnStopAdditionalInfo)
			fmt.Println(response.GetVpnStatus().GetAdditionalInfo())
		}
	}

	return nil
}
