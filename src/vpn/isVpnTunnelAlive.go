package vpn

import (
	"context"
	"net"

	"github.com/zerops-io/zcli/src/nettools"
)

func (h *Handler) isVpnTunnelAlive(serverIp net.IP) bool {

	if serverIp.String() == "" {
		return false
	}

	for i := 0; i < h.config.VpnCheckRetryCount; i++ {
		if func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), h.config.VpnCheckTimeout)
			defer cancel()

			err := nettools.Ping(ctx, serverIp.String())
			if err != nil {
				return false
			}
			return true
		}() {
			return true
		}
	}
	return false
}
