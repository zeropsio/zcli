package statusVpn

import (
	"context"

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
	if err := utils.HandleDaemonError(err); err != nil {
		return err
	}

	utils.PrintVpnStatus(response.GetVpnStatus())
	return nil
}
