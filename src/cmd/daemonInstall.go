package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cliAction/installDaemon"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/daemonInstaller"
	"github.com/zeropsio/zcli/src/i18n"
)

func daemonInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        i18n.CmdDaemonInstall,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
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

	cmd.PersistentFlags().BoolP("help", "h", false, helpText(i18n.DaemonInstallHelp))
	return cmd
}
