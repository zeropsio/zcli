//go:build windows

package vpn

import (
	"context"
	"errors"
	"os"

	"github.com/zeropsio/zcli/src/i18n"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func (h *Handler) cleanVpn(ctx context.Context, interfaceName string) error {

	wg, err := wgctrl.New()
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	defer wg.Close()

	h.logger.Debug("check interface: ", interfaceName)
	if _, err := wg.Device(interfaceName); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return h.runCommands(
		ctx,
		makeCommand(
			"wireguard",
			commandWithErrorMessage(i18n.VpnStopUnableToRemoveTunnelInterface),
			commandWithArgs("/uninstalltunnelservice", interfaceName),
		),
	)
}
