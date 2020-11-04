package utils

import (
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func PrintVpnStatus(vpnStatus *zeropsDaemonProtocol.VpnStatus) {
	if vpnStatus.GetTunnelState() == zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE {
		fmt.Println(i18n.VpnStatusTunnelStatusActive)
		if vpnStatus.GetDnsState() == zeropsDaemonProtocol.DnsState_DNS_ACTIVE {
			fmt.Println(i18n.VpnStatusDnsStatusActive)
		} else {
			fmt.Println(i18n.VpnStatusDnsStatusInactive)
		}
	} else {
		fmt.Println(i18n.VpnStatusTunnelStatusInactive)
	}
	if vpnStatus.GetAdditionalInfo() != "" {
		fmt.Println(i18n.VpnStatusAdditionalInfo)
		fmt.Println(vpnStatus.GetAdditionalInfo())
	}
}
