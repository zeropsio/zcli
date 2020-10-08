package stopVpn

import (
	"context"
	"fmt"

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

	_, err := h.zeropsDaemonClient.StopVpn(ctx, &zeropsDaemonProtocol.StopVpnRequest{})
	if err != nil {
		return err
	}

	fmt.Println(i18n.VpnStopSuccess)

	return nil
}
