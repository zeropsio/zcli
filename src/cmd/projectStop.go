package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"

	"github.com/zeropsio/zcli/src/i18n"
)

func projectStopCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("stop").
		Short(i18n.T(i18n.CmdProjectStop)).
		ScopeLevel(scope.Project).
		Arg(scope.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.ProjectStopHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			stopProjectResponse, err := cmdData.RestApiClient.PutProjectStop(
				ctx,
				path.ProjectId{
					Id: cmdData.Project.ID,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := stopProjectResponse.Output()
			if err != nil {
				return err
			}

			processId := responseOutput.Id

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ProjectStopping),
					ErrorMessageMessage: i18n.T(i18n.ProjectStopping),
					SuccessMessage:      i18n.T(i18n.ProjectStopped),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
