package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
)

func vpnKeyCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("key").
		Short("VPN keys commands group.").
		HelpFlag("Help for the 'vpn key' commands.").
		AddChildrenCmd(vpnKeyListCmd()).
		AddChildrenCmd(vpnKeyRemoveCmd())
}
