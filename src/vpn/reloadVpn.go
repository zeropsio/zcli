package vpn

import (
	"context"
	"errors"
	"os"

	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/i18n"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func (h *Handler) ReloadVpn(
	ctx context.Context,
) (err error) {
	data := h.storage.Data()
	if data.InterfaceName == "" {
		return nil
	}
	wg, err := wgctrl.New()
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	defer wg.Close()

	h.logger.Debug("check Interface: ", data.InterfaceName)
	if _, err := wg.Device(data.InterfaceName); err != nil {
		if os.IsNotExist(err) {
			return h.stopVpn(ctx)
		}
		return err
	}
	return dns.ReloadDns(data, h.dnsServer)
}
