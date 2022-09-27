package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/i18n"
)

func projectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: i18n.CmdProject}

	cmd.AddCommand(projectStartCmd(), projectStopCmd(), projectDeleteCmd(), projectImportCmd())
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))
	return cmd
}
