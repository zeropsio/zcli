package vpn

import (
	"context"
	"net"
	"os/exec"
	"runtime"
)

func (h *Handler) isVpnTunnelAlive(serverIp net.IP) bool {

	if serverIp.String() == "" {
		return false
	}

	for i := 0; i < h.config.VpnCheckRetryCount; i++ {
		if func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), h.config.VpnCheckTimeout)
			defer cancel()

			pingCommand := exec.CommandContext(ctx, "ping6", "-c", "1", serverIp.String())
			if runtime.GOOS == "windows" {
				pingCommand = exec.CommandContext(ctx, "ping", "/n", "1", serverIp.String())
			}

			_, err := pingCommand.Output()
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
