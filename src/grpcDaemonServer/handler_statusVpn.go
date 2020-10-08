package grpcDaemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StatusVpn(ctx context.Context, request *zeropsDaemonProtocol.StatusVpnRequest) (*zeropsDaemonProtocol.StatusVpnResponse, error) {

	alive := h.vpn.StatusVpn()
	if alive {
		return &zeropsDaemonProtocol.StatusVpnResponse{
			Status: zeropsDaemonProtocol.VpnStatus_ACTIVE,
			Error:  nil,
		}, nil
	}

	return &zeropsDaemonProtocol.StatusVpnResponse{
		Status: zeropsDaemonProtocol.VpnStatus_INACTIVE,
		Error:  nil,
	}, nil
}
