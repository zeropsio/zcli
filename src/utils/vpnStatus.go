package utils

import (
	"fmt"

	"github.com/zeropsio/zcli/src/proto/daemon"

	"github.com/zeropsio/zcli/src/i18n"
)

func PrintVpnStatus(vpnStatus *daemon.VpnStatus) {
	switch vpnStatus.GetTunnelState() {
	case daemon.TunnelState_TUNNEL_ACTIVE:
		fmt.Println(i18n.VpnStatusTunnelStatusActive)
	case daemon.TunnelState_TUNNEL_SET_INACTIVE:
		fmt.Println(i18n.VpnStatusTunnelStatusSetInactive)
	}

	if vpnStatus.GetTunnelState() == daemon.TunnelState_TUNNEL_ACTIVE {
		switch vpnStatus.GetDnsState() {
		case daemon.DnsState_DNS_ACTIVE:
			fmt.Println(i18n.VpnStatusDnsStatusActive)
		case daemon.DnsState_DNS_SET_INACTIVE:
			fmt.Println(i18n.VpnStatusDnsStatusSetInactive)
		}
	}
	if vpnStatus.GetAdditionalInfo() != "" {
		fmt.Println(i18n.VpnStatusAdditionalInfo)
		fmt.Println(vpnStatus.GetAdditionalInfo())
	}
}
