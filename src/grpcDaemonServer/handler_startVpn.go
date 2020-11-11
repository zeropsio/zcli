package grpcDaemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StartVpn(ctx context.Context, request *zeropsDaemonProtocol.StartVpnRequest) (*zeropsDaemonProtocol.StartVpnResponse, error) {

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
		return &zeropsDaemonProtocol.StartVpnResponse{}, err
	}

	return &zeropsDaemonProtocol.StartVpnResponse{
		VpnStatus: vpnStatus,
	}, nil
}
