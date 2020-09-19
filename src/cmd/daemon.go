package cmd

import (
	"github.com/zerops-io/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func daemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "daemon",
		Short:        i18n.CmdDaemon,
		SilenceUsage: true,
	}

	cmd.AddCommand(daemonRunCmd())
	cmd.AddCommand(daemonInstallCmd())
	cmd.AddCommand(daemonRemoveCmd())

	return cmd
}
