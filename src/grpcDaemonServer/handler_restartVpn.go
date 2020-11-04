package grpcDaemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) RestartVpn(ctx context.Context, request *zeropsDaemonProtocol.RestartVpnRequest) (*zeropsDaemonProtocol.RestartVpnResponse, error) {

	vpnStatus, err := h.vpn.RestartVpn(ctx, request.GetToken())
	if err != nil {
		return &zeropsDaemonProtocol.RestartVpnResponse{}, err
	}

	return &zeropsDaemonProtocol.RestartVpnResponse{
		VpnStatus: vpnStatus,
	}, nil
}
