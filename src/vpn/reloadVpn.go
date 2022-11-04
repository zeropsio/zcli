package vpn

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/zeropsio/zcli/src/i18n"
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

	h.logger.Debugf("check vpn interface %s", data.InterfaceName)
	if _, err := wg.Device(data.InterfaceName); err != nil {
		if os.IsNotExist(err) {
			h.logger.Debugf("interface %s not exists, starting vpn", data.InterfaceName)
			if err := h.DnsClean(ctx); err != nil {
				return err
			}
			time.Sleep(time.Second * 3)
			return h.startVpn(
				ctx,
				data.GrpcApiAddress,
				data.GrpcVpnAddress,
				data.Token,
				data.ProjectId,
				data.UserId,
				data.Mtu,
				data.CaCertificateUrl,
				data.PreferredPortMin,
				data.PreferredPortMax,
			)
		}
		return err
	}
	if err := h.DnsClean(ctx); err != nil {
		return err
	}
	if err := h.setDns(ctx); err != nil {
		return err
	}
	return nil
}
