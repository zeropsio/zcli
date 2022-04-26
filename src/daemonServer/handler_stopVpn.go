package daemonServer

import (
	"context"
	"github.com/zerops-io/zcli/src/proto/daemon"
)

func (h *Handler) StopVpn(ctx context.Context, request *daemon.StopVpnRequest) (*daemon.StopVpnResponse, error) {
	return h.vpn.StopVpn()
}
