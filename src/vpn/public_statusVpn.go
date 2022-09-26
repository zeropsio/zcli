package vpn

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/proto/daemon"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/dns"
)

func (h *Handler) StatusVpn(ctx context.Context) (*daemon.VpnStatus, error) {
	data := h.storage.Data()

	vpnStatus := &daemon.VpnStatus{
		TunnelState: daemon.TunnelState_TUNNEL_UNSET,
		DnsState:    daemon.DnsState_DNS_UNSET,
	}

	if data.ServerIp != nil {
		vpnStatus.TunnelState = daemon.TunnelState_TUNNEL_SET_INACTIVE
	}
	if data.DnsManagement != daemonStorage.LocalDnsManagementUnknown {
		vpnStatus.DnsState = daemon.DnsState_DNS_SET_INACTIVE
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

	h.logger.Debug("check Interface: ", data.InterfaceName)
	if _, err := wg.Device(data.InterfaceName); err != nil {
		if os.IsNotExist(err) {
			return vpnStatus, h.stopVpn(ctx)
		}
		return vpnStatus, err
	} else {
		if !h.isVpnTunnelAlive(ctx, data.ServerIp) {
			vpnStatus.TunnelState = daemon.TunnelState_TUNNEL_SET_INACTIVE
			return vpnStatus, nil
		}
	}
	vpnStatus.TunnelState = daemon.TunnelState_TUNNEL_ACTIVE

	if vpnStatus.DnsState == daemon.DnsState_DNS_SET_INACTIVE {
		dnsIsAlive, err := dns.IsAlive()
		if err != nil {
			h.logger.Error(err)
			vpnStatus.AdditionalInfo = i18n.VpnStatusDnsCheckError + "\n"
		}
		if dnsIsAlive {
			vpnStatus.DnsState = daemon.DnsState_DNS_ACTIVE
		}

		if vpnStatus.DnsState != daemon.DnsState_DNS_ACTIVE {
			vpnStatus.AdditionalInfo += fmt.Sprintf(
				"dns ip: %s\n"+
					"vpn network: %s\n"+
					"client ip: %s\n"+
					"interface: %s\n",
				data.DnsIp.String(),
				data.VpnNetwork.String(),
				data.ClientIp.String(),
				data.InterfaceName,
			)
		}
	}

	return vpnStatus, nil
}
