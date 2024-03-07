package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func vpnCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("vpn").
		Short(i18n.T(i18n.CmdVpn)).
		HelpFlag(i18n.T(i18n.VpnHelp)).
		AddChildrenCmd(vpnConnectCmd()).
		AddChildrenCmd(vpnDisconnectCmd())
}
