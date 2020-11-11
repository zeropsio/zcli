package utils

import (
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func PrintVpnStatus(vpnStatus *zeropsDaemonProtocol.VpnStatus) {
	switch vpnStatus.GetTunnelState() {
	case zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE:
		fmt.Println(i18n.VpnStatusTunnelStatusActive)
	case zeropsDaemonProtocol.TunnelState_TUNNEL_SET_INACTIVE:
		fmt.Println(i18n.VpnStatusTunnelStatusSetInactive)
	case zeropsDaemonProtocol.TunnelState_TUNNEL_UNSET:
		fmt.Println(i18n.VpnStatusTunnelStatusUnset)
	}

	if vpnStatus.GetTunnelState() == zeropsDaemonProtocol.TunnelState_TUNNEL_ACTIVE {
		switch vpnStatus.GetDnsState() {
		case zeropsDaemonProtocol.DnsState_DNS_ACTIVE:
			fmt.Println(i18n.VpnStatusDnsStatusActive)
		case zeropsDaemonProtocol.DnsState_DNS_SET_INACTIVE:
			fmt.Println(i18n.VpnStatusDnsStatusSetInactive)
		case zeropsDaemonProtocol.DnsState_DNS_UNSET:
			fmt.Println(i18n.VpnStatusDnsStatusUnset)
		}
	}
	if vpnStatus.GetAdditionalInfo() != "" {
		fmt.Println(i18n.VpnStatusAdditionalInfo)
		fmt.Println(vpnStatus.GetAdditionalInfo())
	}
}
