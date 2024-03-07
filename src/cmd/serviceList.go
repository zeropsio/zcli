package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func serviceListCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("list").
		Short(i18n.T(i18n.CmdServiceList)).
		ScopeLevel(scope.Project).
		Arg(scope.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.ServiceListHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			err := uxHelpers.PrintServiceList(ctx, cmdData.UxBlocks, cmdData.RestApiClient, *cmdData.Project)
			if err != nil {
				return err
			}

			return nil
		})
}
