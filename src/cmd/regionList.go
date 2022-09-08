package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/i18n"
)

const defaultRegionUrl = "https://api.app.zerops.io/api/rest/public/region/zcli"

func regionList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: i18n.CmdRegion,
	}

	listCmd := &cobra.Command{
		Use:          "list",
		Short:        i18n.CmdRegionList,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			regSignals(cancel)

			region, err := createRegionRetriever(ctx)
			if err != nil {
				return err
			}

			regionURL := params.GetString(cmd, "regionURL")
			regions, err := region.RetrieveAllFromURL(regionURL)
			if err != nil {
				return err
			}

			for _, r := range regions {
				fmt.Print(r.Name)
				if r.IsDefault {
					fmt.Print(" [default]")
				}
				fmt.Println()
			}
			return nil
		},
	}
	params.RegisterString(listCmd, "regionURL", defaultRegionUrl, "zerops region")
	listCmd.Flags().BoolP("help", "h", false, helpText(i18n.RegionListHelp))

	listCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		if err := command.Flags().MarkHidden("regionURL"); err != nil {
			return
		}
		command.Parent().HelpFunc()(command, strings)
	})
	cmd.AddCommand(listCmd)

	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))
	return cmd
}
