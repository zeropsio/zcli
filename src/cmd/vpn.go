package cmd

import (
	"github.com/zerops-io/zcli/src/i18n"

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

	return cmd
}
