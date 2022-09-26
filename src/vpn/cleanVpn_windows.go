//go:build windows

package vpn

import (
	"context"
	"errors"
	"os"

	"github.com/zerops-io/zcli/src/i18n"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func (h *Handler) cleanVpn(ctx context.Context, interfaceName string) error {

	wg, err := wgctrl.New()
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	defer wg.Close()

	h.logger.Debug("check Interface: ", interfaceName)
	if _, err := wg.Device(interfaceName); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return runCommands(
		ctx,
		h.logger,
		makeCommand(
			"wireguard",
			i18n.VpnStopUnableToRemoveTunnelInterface,
			"/uninstalltunnelservice", interfaceName,
		),
	)
}
