package cmd

import (
	"github.com/zeropsio/zcli/src/i18n"

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

	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))

	return cmd
}
