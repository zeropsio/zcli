package vpn

import (
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) stopVpn() (vpnStatus *zeropsDaemonProtocol.VpnStatus, err error) {
	data := h.storage.Data()

	err = h.cleanVpn()
	if err != nil {
		return nil, err
	}

	localDnsManagement, err := dns.DetectDns()
	if err != nil {
		return nil, err
	}

	if data.InterfaceName != "" {
		err = dns.CleanDns(h.dnsServer, data.DnsIp, data.InterfaceName, localDnsManagement)
		if err != nil {
			return nil, err
		}
	}

	dataReset := &daemonStorage.Data{}
	err = h.storage.Save(dataReset)
	if err != nil {
		return nil, err
	}

	vpnStatus = &zeropsDaemonProtocol.VpnStatus{
		TunnelState: zeropsDaemonProtocol.TunnelState_TUNNEL_UNSET,
		DnsState:    zeropsDaemonProtocol.DnsState_DNS_UNSET,
	}

	if localDnsManagement == dns.LocalDnsManagementUnknown {
		vpnStatus.AdditionalInfo = i18n.VpnStopAdditionalInfoMessage
	}

	return
}
