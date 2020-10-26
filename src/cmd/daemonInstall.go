package cmd

import (
	"context"

	"github.com/zerops-io/zcli/src/cliAction/installDaemon"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/i18n"
)

func daemonInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        i18n.CmdDaemonInstall,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			installer, err := daemonInstaller.New(daemonInstaller.Config{})
			if err != nil {
				return err
			}

			return installDaemon.New(
				installDaemon.Config{},
				installer,
			).
				Run(ctx, installDaemon.RunConfig{})
		},
	}

	return cmd
}
