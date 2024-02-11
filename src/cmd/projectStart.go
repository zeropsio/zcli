package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"
)

func projectStartCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("start").
		Short(i18n.T(i18n.CmdProjectStart)).
		ScopeLevel(cmdBuilder.Project).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.ProjectStartHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			startProjectResponse, err := cmdData.RestApiClient.PutProjectStart(
				ctx,
				path.ProjectId{
					Id: cmdData.Project.ID,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := startProjectResponse.Output()
			if err != nil {
				return err
			}

			processId := responseOutput.Id

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ProjectStarting),
					ErrorMessageMessage: i18n.T(i18n.ProjectStarting),
					SuccessMessage:      i18n.T(i18n.ProjectStarted),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
