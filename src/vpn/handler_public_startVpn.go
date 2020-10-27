package vpn

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StartVpn(
	ctx context.Context,
	grpcApiAddress string,
	grpcVpnAddress string,
	token string,
	projectId string,
	mtu uint32,
) (vpnStatus *zeropsDaemonProtocol.VpnStatus, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	err = h.startVpn(ctx, grpcApiAddress, grpcVpnAddress, token, projectId, mtu)
	if err != nil {
		return nil, err
	}
	vpnStatus = &zeropsDaemonProtocol.VpnStatus{
		TunnelState: zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE,
		DnsState:    zeropsDaemonProtocol.DnsState_DNS_ACTIVE,
	}

	data := h.storage.Data()
	if data.DnsManagement == string(localDnsManagementUnknown) {
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
