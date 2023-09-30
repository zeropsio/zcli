package vpn

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/daemon"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func (h *Handler) StatusVpn(ctx context.Context) (*daemon.VpnStatus, error) {
	data := h.storage.Data()

	vpnStatus := &daemon.VpnStatus{
		TunnelState: daemon.TunnelState_TUNNEL_SET_INACTIVE,
		DnsState:    daemon.DnsState_DNS_SET_INACTIVE,
	}

	if data.InterfaceName == "" {
		return vpnStatus, nil
	}

	wg, err := wgctrl.New()
	if err != nil {
		h.logger.Error(err)
		return nil, errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	defer wg.Close()

	h.logger.Debug("check interface: ", data.InterfaceName)
	if _, err := wg.Device(data.InterfaceName); err != nil {
		if os.IsNotExist(err) {
			return vpnStatus, h.stopVpn(ctx)
		}
		return vpnStatus, err
	} else {
		if !h.isVpnTunnelAlive(ctx, data.ServerIp) {
			return vpnStatus, nil
		}
	}
	vpnStatus.TunnelState = daemon.TunnelState_TUNNEL_ACTIVE

	dnsIsAlive, err := h.dnsIsAlive()
	if err != nil {
		h.logger.Error(err)
		vpnStatus.AdditionalInfo = i18n.VpnStatusDnsCheckError + "\n"
	}
	if dnsIsAlive {
		vpnStatus.DnsState = daemon.DnsState_DNS_ACTIVE
	} else {
		vpnStatus.AdditionalInfo += fmt.Sprintf(
			"dns ip: %s, %s\n"+
				"vpn network: %s, %s\n"+
				"client ip: %s\n"+
				"interface: %s\n",
			data.DnsIp.String(), data.DnsIp4.String(),
			data.VpnNetwork.String(), data.VpnNetwork4.String(),
			data.ClientIp.String(), data.ClientIp4.String(),
			data.InterfaceName,
		)
	}
	return vpnStatus, nil
}
