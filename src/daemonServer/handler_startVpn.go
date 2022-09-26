package daemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/proto/daemon"
)

func (h *Handler) StartVpn(ctx context.Context, request *daemon.StartVpnRequest) (*daemon.VpnStatus, error) {

	if err := h.vpn.StartVpn(
		ctx,
		request.GetApiAddress(),
		request.GetVpnAddress(),
		request.GetToken(),
		request.GetProjectId(),
		request.GetUserId(),
		request.GetMtu(),
		request.GetCaCertificateUrl(),
		request.GetPreferredPortMin(),
		request.GetPreferredPortMax(),
	); err != nil {
		return nil, err
	}

	return h.StatusVpn(ctx, &daemon.StatusVpnRequest{})
}
