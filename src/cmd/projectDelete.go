package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"

	"github.com/zeropsio/zcli/src/i18n"
)

func projectDeleteCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("delete").
		Short(i18n.T(i18n.CmdProjectDelete)).
		ScopeLevel(cmdBuilder.Project).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			confirm, err := YesNoPromptDestructive(ctx, cmdData, i18n.T(i18n.ProjectDeleteConfirm, cmdData.Project.Name))
			if err != nil {
				return err
			}

			if !confirm {
				// TODO - janhajek message
				fmt.Println("you have to confirm it")
				return nil
			}

			deleteProjectResponse, err := cmdData.RestApiClient.DeleteProject(
				ctx,
				path.ProjectId{
					Id: cmdData.Project.ID,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := deleteProjectResponse.Output()
			if err != nil {
				return err
			}

			processId := responseOutput.Id

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				cmdData.RestApiClient,
				[]uxHelpers.Process{{
					Id:                  processId,
					RunningMessage:      i18n.T(i18n.ProjectDeleting),
					ErrorMessageMessage: i18n.T(i18n.ProjectDeleting),
					SuccessMessage:      i18n.T(i18n.ProjectDeleted),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
