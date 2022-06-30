package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/i18n"
)

func serviceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "service", Short: i18n.CmdService}

	cmd.AddCommand(serviceStartCmd(), serviceStopCmd(), serviceDeleteCmd(), serviceImportCmd(), serviceLogCmd())
	return cmd
}
