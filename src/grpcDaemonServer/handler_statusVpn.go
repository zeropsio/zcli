package grpcDaemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StatusVpn(ctx context.Context, request *zeropsDaemonProtocol.StatusVpnRequest) (*zeropsDaemonProtocol.StatusVpnResponse, error) {
	vpnStatus := h.vpn.StatusVpn()

	return &zeropsDaemonProtocol.StatusVpnResponse{
		VpnStatus: vpnStatus,
	}, nil
}
