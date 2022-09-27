package cmd

import (
	"github.com/zeropsio/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func vpnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "vpn",
		Short:        i18n.CmdVpn,
		SilenceUsage: true,
	}

	cmd.AddCommand(vpnStartCmd())
	cmd.AddCommand(vpnStopCmd())
	cmd.AddCommand(vpnStatusCmd())

	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))
	return cmd
}
