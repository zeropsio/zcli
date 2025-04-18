package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func serviceListCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("list").
		Short(i18n.T(i18n.CmdDescServiceList)).
		ScopeLevel(cmdBuilder.Project()).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpServiceList)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}
			if err := uxHelpers.PrintServiceList(
				ctx,
				cmdData.RestApiClient,
				cmdData.Stdout,
				project,
			); err != nil {
				return err
			}
			return nil
		})
}
