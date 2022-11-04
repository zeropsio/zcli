//go:build darwin

package vpn

import (
	"context"
	"errors"
	"net"
	"os"
	"path"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) cleanVpn(_ context.Context, interfaceName string) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStopUnableToRemoveTunnelInterface)
	}
	for _, in := range interfaces {
		if in.Name == interfaceName {
			socket := path.Join("/var/run/wireguard/", interfaceName+".sock")
			h.logger.Debugf("remove wireguard socket: %s", socket)
			if err := os.Remove(socket); err != nil {
				return errors.New(i18n.VpnStopUnableToRemoveTunnelInterface)
			}
		}
	}
	return nil
}
