package vpn

import (
	"context"
	"net"
	"os/exec"
)

func (h *Handler) isVpnAlive(serverIp net.IP) bool {

	if serverIp.String() == "" {
		return false
	}

	for i := 0; i < h.config.VpnCheckRetryCount; i++ {
		if func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), h.config.VpnCheckTimeout)
			defer cancel()
			_, err := exec.CommandContext(ctx, "ping6", "-c", "1", serverIp.String()).Output()
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
