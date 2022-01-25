package vpn

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StartVpn(
	ctx context.Context,
	grpcApiAddress string,
	grpcVpnAddress string,
	token string,
	projectId string,
	userId string,
	mtu uint32,
	caCertificateUrl string,
) (vpnStatus *zeropsDaemonProtocol.VpnStatus, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	err = h.startVpn(
		ctx,
		grpcApiAddress,
		grpcVpnAddress,
		token,
		projectId,
		userId,
		mtu,
		caCertificateUrl,
	)
	if err != nil {
		return nil, err
	}

	// tunnel status was checked in internal function startVpn
	vpnStatus = &zeropsDaemonProtocol.VpnStatus{
		TunnelState: zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE,
		DnsState:    zeropsDaemonProtocol.DnsState_DNS_UNSET,
	}

	data := h.storage.Data()
	if data.DnsManagement != string(dns.LocalDnsManagementUnknown) {
		vpnStatus.DnsState = zeropsDaemonProtocol.DnsState_DNS_SET_INACTIVE
		dnsIsAlive, err := dns.IsAlive()
		if err != nil {
			h.logger.Error(err)
			vpnStatus.AdditionalInfo = i18n.VpnStatusDnsCheckError + "\n"
		}
		if dnsIsAlive {
			vpnStatus.DnsState = zeropsDaemonProtocol.DnsState_DNS_ACTIVE
		}
		if vpnStatus.DnsState != zeropsDaemonProtocol.DnsState_DNS_ACTIVE {
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
