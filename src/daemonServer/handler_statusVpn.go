package daemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/proto/daemon"
)

func (h *Handler) StatusVpn(ctx context.Context, request *daemon.StatusVpnRequest) (*daemon.StatusVpnResponse, error) {
	vpnStatus := h.vpn.StatusVpn()

	return &daemon.StatusVpnResponse{
		VpnStatus: vpnStatus,
	}, nil
}
