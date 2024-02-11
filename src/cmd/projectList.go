package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func projectListCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("list").
		Short(i18n.T(i18n.CmdProjectList)).
		HelpFlag(i18n.T(i18n.ProjectListHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			err := uxHelpers.PrintProjectList(ctx, cmdData.UxBlocks, cmdData.RestApiClient)
			if err != nil {
				return err
			}

			return nil
		})
}
