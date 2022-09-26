//go:build linux

package vpn

import (
	"context"
	"net"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) cleanVpn(ctx context.Context, interfaceName string) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, in := range interfaces {
		if in.Name == interfaceName {
			return runCommands(
				ctx,
				h.logger,
				makeCommand(
					"ip",
					i18n.VpnStopUnableToRemoveTunnelInterface,
					"link", "del", interfaceName,
				),
			)
		}
	}
	return nil
}
