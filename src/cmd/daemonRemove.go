package cmd

import (
	"context"
	"github.com/zerops-io/zcli/src/proto/daemon"

	"github.com/zerops-io/zcli/src/cliAction/removeDaemon"
	"github.com/zerops-io/zcli/src/cliAction/stopVpn"
	"github.com/zerops-io/zcli/src/daemonInstaller"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func daemonRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "remove",
		Short:        i18n.CmdDaemonRemove,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			daemonClient, daemonCloseFunc, err := daemon.CreateClient(ctx)
			if err != nil {
				return err
			}
			defer daemonCloseFunc()

			installer, err := daemonInstaller.New(daemonInstaller.Config{})
			if err != nil {
				return err
			}

			stopVpn := stopVpn.New(
				stopVpn.Config{},
				daemonClient,
			)

			return removeDaemon.New(
				removeDaemon.Config{},
				installer,
				stopVpn,
			).
				Run(ctx, removeDaemon.RunConfig{})
		},
	}

	return cmd
}
