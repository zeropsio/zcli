package cmd

import (
	"context"
	"github.com/zerops-io/zcli/src/proto/daemon"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/cliAction/statusVpn"

	"github.com/spf13/cobra"
)

func vpnStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        i18n.CmdVpnStatus,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
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

	return cmd
}
