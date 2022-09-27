package daemonServer

import (
	"context"

	"github.com/zeropsio/zcli/src/proto/daemon"
)

func (h *Handler) StopVpn(ctx context.Context, _ *daemon.StopVpnRequest) (*daemon.VpnStatus, error) {
	if err := h.vpn.StopVpn(ctx); err != nil {
		return nil, err
	}
	return h.vpn.StatusVpn(ctx)
}
