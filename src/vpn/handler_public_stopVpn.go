package vpn

import (
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StopVpn() (vpnStatus *zeropsDaemonProtocol.VpnStatus, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	defer func() {
		if err != nil {
			h.logger.Error(err)
		}
	}()

	err = h.cleanVpn()
	if err != nil {
		return nil, err
	}

	localDnsManagement, err := h.detectDns()
	if err != nil {
		return nil, err
	}

	data := h.storage.Data()
	err = h.cleanDns(data.DnsIp, data.ClientIp, localDnsManagement)
	if err != nil {
		return nil, err
	}

	dataReset := &daemonStorage.Data{}
	err = h.storage.Save(dataReset)
	if err != nil {
		return nil, err
	}

	vpnStatus = &zeropsDaemonProtocol.VpnStatus{
		TunnelState: zeropsDaemonProtocol.TunnelState_TUNNEL_INACTIVE,
		DnsState:    zeropsDaemonProtocol.DnsState_DNS_INACTIVE,
	}

	if localDnsManagement == localDnsManagementUnknown {
		vpnStatus.AdditionalInfo = i18n.VpnStopAdditionalInfoMessage
	}

	return
}
