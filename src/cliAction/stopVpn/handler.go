package stopVpn

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

	response, err := h.zeropsDaemonClient.StopVpn(ctx, &zeropsDaemonProtocol.StopVpnRequest{})
	if err := utils.HandleDaemonError(err); err != nil {
		return err
	}

	fmt.Println(i18n.VpnStopSuccess)
	status := response.GetVpnStatus()
	if status.GetAdditionalInfo() != "" {
		fmt.Println(i18n.VpnStopAdditionalInfo)
		fmt.Println(status.GetAdditionalInfo())
	}

	return nil
}
