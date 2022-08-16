package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/i18n"
)

func bucketZeropsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "zerops", Short: i18n.CmdBucketZerops}

	cmd.AddCommand(bucketZeropsCreateCmd(), bucketZeropsDeleteCmd())
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))
	return cmd
}
