package stopVpn

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/daemon"

	"github.com/zeropsio/zcli/src/i18n"
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
		fmt.Println(i18n.VpnDaemonUnavailable)
		return nil
	}

	if response.GetTunnelState() == daemon.TunnelState_TUNNEL_SET_INACTIVE {
		fmt.Println(i18n.VpnStopSuccess)
		if response.GetAdditionalInfo() != "" {
			fmt.Println(i18n.VpnStopAdditionalInfo)
			fmt.Println(response.GetAdditionalInfo())
		}
	}

	return nil
}
