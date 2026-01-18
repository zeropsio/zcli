package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func projectProcessesCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("processes").
		Short(i18n.T(i18n.CmdDescProjectProcesses)).
		Long(i18n.T(i18n.CmdDescProjectProcessesLong)).
		ScopeLevel(cmdBuilder.ScopeProject()).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpProjectProcesses)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			return uxHelpers.PrintProcessList(
				ctx,
				cmdData.RestApiClient,
				cmdData.Stdout,
				project.OrgId,
				project.Id,
			)
		})
}
