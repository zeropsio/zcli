package daemonServer

import (
	"context"

	"github.com/zeropsio/zcli/src/proto/daemon"
)

func (h *Handler) StatusVpn(ctx context.Context, _ *daemon.StatusVpnRequest) (*daemon.VpnStatus, error) {
	return h.vpn.StatusVpn(ctx)
}
