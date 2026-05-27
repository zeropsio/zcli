package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/output"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func serviceListCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("list").
		Short(i18n.T(i18n.CmdDescServiceList)).
		ScopeLevel(cmdBuilder.ScopeProject()).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		StringFlag("output", "table", i18n.T(i18n.OutputFormatFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpServiceList)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			outputFormat, err := output.ParseFormat(cmdData.Params.GetString("output"))
			if err != nil {
				return err
			}

			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}
			if err := uxHelpers.PrintServiceList(
				ctx,
				cmdData.RestApiClient,
				cmdData.Stdout,
				project,
				outputFormat,
			); err != nil {
				return err
			}
			return nil
		})
}
