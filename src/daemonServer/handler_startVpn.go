package daemonServer

import (
	"context"
	"github.com/zerops-io/zcli/src/proto/daemon"
)

func (h *Handler) StartVpn(ctx context.Context, request *daemon.StartVpnRequest) (*daemon.StartVpnResponse, error) {

	vpnStatus, err := h.vpn.StartVpn(
		ctx,
		request.GetApiAddress(),
		request.GetVpnAddress(),
		request.GetToken(),
		request.GetProjectId(),
		request.GetUserId(),
		request.GetMtu(),
		request.GetCaCertificateUrl(),
	)
	if err != nil {
		return &daemon.StartVpnResponse{}, err
	}

	return &daemon.StartVpnResponse{
		VpnStatus: vpnStatus,
	}, nil
}
