package vpn

import (
	"fmt"

	"github.com/zerops-io/zcli/src/dns"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StatusVpn() (vpnStatus *zeropsDaemonProtocol.VpnStatus) {
	h.lock.Lock()
	defer h.lock.Unlock()

	data := h.storage.Data()

	if !h.isVpnAlive(data.ServerIp) {
		return
	}

	vpnStatus = &zeropsDaemonProtocol.VpnStatus{
		TunnelState: zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE,
		DnsState:    zeropsDaemonProtocol.DnsState_DNS_ACTIVE,
	}

	if h.storage.Data().DnsManagement == string(dns.LocalDnsManagementUnknown) {
		vpnStatus.DnsState = zeropsDaemonProtocol.DnsState_DNS_INACTIVE
		vpnStatus.AdditionalInfo = fmt.Sprintf(
			"dns ip: %s\n"+
				"vpn network: %s\n"+
				"client ip: %s\n",
			data.DnsIp.String(),
			data.VpnNetwork.String(),
			data.ClientIp.String(),
		)
	}

	return
}
