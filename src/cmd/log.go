package cmd

import (
	"github.com/zerops-io/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func logCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "log",
		Short:        i18n.CmdLog,
		SilenceUsage: true,
	}

	cmd.AddCommand(logShowCmd())

	return cmd
}
