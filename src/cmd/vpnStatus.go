package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/proto/daemon"

	"github.com/zeropsio/zcli/src/i18n"

	"github.com/zeropsio/zcli/src/cliAction/statusVpn"

	"github.com/spf13/cobra"
)

func vpnStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        i18n.CmdVpnStatus,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			regSignals(cancel)

			daemonClient, daemonCloseFunc, err := daemon.CreateClient(ctx)
			if err != nil {
				return err
			}
			defer daemonCloseFunc()

			return statusVpn.New(
				statusVpn.Config{},
				daemonClient,
			).Run(ctx, statusVpn.RunConfig{})
		},
	}
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.VpnStatusHelp))
	return cmd
}
