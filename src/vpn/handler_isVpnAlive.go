package vpn

import (
	"context"
	"os/exec"
)

func (h *Handler) isVpnAlive(serverIp string) bool {

	if serverIp == "" {
		return false
	}

	for i := 0; i < h.config.VpnCheckRetryCount; i++ {
		if func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), h.config.VpnCheckTimeout)
			defer cancel()
			_, err := exec.CommandContext(ctx, "ping6", "-c", "1", serverIp).Output()
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
