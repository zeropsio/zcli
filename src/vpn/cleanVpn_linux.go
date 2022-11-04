//go:build linux

package vpn

import (
	"context"
	"net"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) cleanVpn(ctx context.Context, interfaceName string) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, in := range interfaces {
		if in.Name == interfaceName {
			return h.runCommands(
				ctx,
				makeCommand(
					"ip",
					commandWithErrorMessage(i18n.VpnStopUnableToRemoveTunnelInterface),
					commandWithArgs("link", "del", interfaceName),
				),
			)
		}
	}
	return nil
}
