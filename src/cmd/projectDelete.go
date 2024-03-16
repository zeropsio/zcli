package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"
)

func projectDeleteCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("delete").
		Short(i18n.T(i18n.CmdProjectDelete)).
		ScopeLevel(scope.Project).
		Arg(scope.ProjectArgName, cmdBuilder.OptionalArg()).
		BoolFlag("confirm", false, i18n.T(i18n.ConfirmFlag)).
		HelpFlag(i18n.T(i18n.ProjectDeleteHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			if !cmdData.Params.GetBool("confirm") {
				confirmed, err := uxHelpers.YesNoPrompt(
					ctx,
					cmdData.UxBlocks,
					i18n.T(i18n.ProjectDeleteConfirm, cmdData.Project.Name),
				)
				if err != nil {
					return err
				}
				if !confirmed {
					return errors.New(i18n.T(i18n.DestructiveOperationConfirmationFailed))
				}
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
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
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
