package statusVpn

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/utils"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

type Config struct {
}

type RunConfig struct {
}

type Handler struct {
	config             Config
	zeropsDaemonClient zeropsDaemonProtocol.ZeropsDaemonProtocolClient
}

func New(
	config Config,
	zeropsDaemonClient zeropsDaemonProtocol.ZeropsDaemonProtocolClient,
) *Handler {
	return &Handler{
		config:             config,
		zeropsDaemonClient: zeropsDaemonClient,
	}
}

func (h *Handler) Run(ctx context.Context, _ RunConfig) error {

	response, err := h.zeropsDaemonClient.StatusVpn(ctx, &zeropsDaemonProtocol.StatusVpnRequest{})
	if err := utils.HandleDaemonError(err); err != nil {
		return err
	}

	status := response.GetVpnStatus()
	if status.GetTunnelState() == zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE {
		fmt.Println(i18n.VpnStatusTunnelStatusActive)

		if status.GetDnsState() == zeropsDaemonProtocol.DnsState_DNS_ACTIVE {
			fmt.Println(i18n.VpnStatusDnsStatusActive)
		} else {
			fmt.Println(i18n.VpnStatusDnsStatusInactive)
		}
	} else {
		fmt.Println(i18n.VpnStatusTunnelStatusInactive)
	}

	if status.GetAdditionalInfo() != "" {
		fmt.Println(i18n.VpnStatusAdditionalInfo)
		fmt.Println(status.GetAdditionalInfo())
	}

	return nil
}
