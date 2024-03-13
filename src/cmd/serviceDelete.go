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

func serviceDeleteCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("delete").
		Short(i18n.T(i18n.CmdServiceDelete)).
		ScopeLevel(scope.Service).
		Arg(scope.ServiceArgName, cmdBuilder.OptionalArg()).
		BoolFlag("confirm", false, i18n.T(i18n.ConfirmFlag)).
		HelpFlag(i18n.T(i18n.ServiceDeleteHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			if !cmdData.Params.GetBool("confirm") {
				confirmed, err := uxHelpers.YesNoPrompt(
					ctx,
					cmdData.UxBlocks,
					i18n.T(i18n.ServiceDeleteConfirm, cmdData.Service.Name),
				)
				if err != nil {
					return err
				}
				if !confirmed {
					return errors.New(i18n.T(i18n.DestructiveOperationConfirmationFailed))
				}
			}

			deleteServiceResponse, err := cmdData.RestApiClient.DeleteServiceStack(
				ctx,
				path.ServiceStackId{
					Id: cmdData.Service.ID,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := deleteServiceResponse.Output()
			if err != nil {
				return err
			}

			processId := responseOutput.Id

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ServiceDeleting),
					ErrorMessageMessage: i18n.T(i18n.ServiceDeleting),
					SuccessMessage:      i18n.T(i18n.ServiceDeleted),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
