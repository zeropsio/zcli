package cmd

import (
	"github.com/zeropsio/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func logCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "log",
		Short:        i18n.CmdLog,
		SilenceUsage: true,
	}

	cmd.AddCommand(logShowCmd())
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))

	return cmd
}
