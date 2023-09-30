package vpn

import (
	"context"
	"net"
)

func (h *Handler) isVpnTunnelAlive(ctx context.Context, serverIp net.IP) bool {

	return true
	/*
		if serverIp.String() == "" {
			return false
		}

		for i := 0; i < h.config.VpnCheckRetryCount; i++ {
			if func() bool {
				ctx, cancel := context.WithTimeout(ctx, h.config.VpnCheckTimeout)
				defer cancel()

				err := nettools.Ping(ctx, serverIp.String())
				if err != nil {
					h.logger.Error(err)
					return false
				}
				return true
			}() {
				return true
			}
		}
	*/
}
