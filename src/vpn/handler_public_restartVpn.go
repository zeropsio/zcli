package vpn

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) RestartVpn(ctx context.Context, token string) (*zeropsDaemonProtocol.VpnStatus, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	data := h.storage.Data()
	if !h.isVpnAlive(data.ServerIp) {
		return &zeropsDaemonProtocol.VpnStatus{}, nil
	}

	_, err := h.stopVpn()
	if err != nil {
		return nil, err
	}

	err = h.startVpn(
		ctx,
		data.GrpcApiAddress,
		data.GrpcVpnAddress,
		token,
		data.ProjectId,
		data.UserId,
		data.Mtu,
		data.CaCertificate,
	)
	if err != nil {
		return nil, err
	}
	vpnStatus := &zeropsDaemonProtocol.VpnStatus{
		TunnelState: zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE,
		DnsState:    zeropsDaemonProtocol.DnsState_DNS_ACTIVE,
	}

	data = h.storage.Data()
	if data.DnsManagement == string(dns.LocalDnsManagementUnknown) {
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

	return vpnStatus, nil
}
