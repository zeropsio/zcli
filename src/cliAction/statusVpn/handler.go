package statusVpn

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/utils"

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
	daemonInstalled, err := utils.HandleDaemonError(err)
	if err != nil {
		return err
	}

	if !daemonInstalled {
		fmt.Println(i18n.VpnStatusDaemonIsUnavailable)
		return nil
	}

	utils.PrintVpnStatus(response.GetVpnStatus())
	return nil
}
