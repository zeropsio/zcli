package vpn

import (
	"fmt"
	"github.com/zerops-io/zcli/src/proto/daemon"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/dns"
)

func (h *Handler) StatusVpn() (vpnStatus *daemon.VpnStatus) {
	h.lock.Lock()
	defer h.lock.Unlock()

	data := h.storage.Data()

	vpnStatus = &daemon.VpnStatus{
		TunnelState: daemon.TunnelState_TUNNEL_UNSET,
		DnsState:    daemon.DnsState_DNS_UNSET,
	}

	if data.ServerIp != nil {
		vpnStatus.TunnelState = daemon.TunnelState_TUNNEL_SET_INACTIVE
	}
	if data.DnsManagement != string(dns.LocalDnsManagementUnknown) {
		vpnStatus.DnsState = daemon.DnsState_DNS_SET_INACTIVE
	}

	if !h.isVpnTunnelAlive(data.ServerIp) {
		return
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
					"client ip: %s\n",
				data.DnsIp.String(),
				data.VpnNetwork.String(),
				data.ClientIp.String(),
			)
		}
	}

	return
}
