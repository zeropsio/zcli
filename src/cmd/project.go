package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/i18n"
)

func projectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: i18n.CmdProject}

	cmd.AddCommand(projectStartCmd(), projectStopCmd(), projectDeleteCmd(), projectImportCmd())
	return cmd
}
