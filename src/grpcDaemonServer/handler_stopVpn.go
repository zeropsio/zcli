package grpcDaemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StopVpn(ctx context.Context, request *zeropsDaemonProtocol.StopVpnRequest) (*zeropsDaemonProtocol.StopVpnResponse, error) {
	return h.vpn.StopVpn()
}
