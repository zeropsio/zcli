package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"
)

func serviceDeleteCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("delete").
		Short(i18n.T(i18n.CmdServiceDelete)).
		ScopeLevel(cmdBuilder.Service).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			confirm, err := YesNoPromptDestructive(ctx, cmdData, i18n.T(i18n.ServiceDeleteConfirm, cmdData.Service.Name))
			if err != nil {
				return err
			}

			if !confirm {
				// FIXME - janhajek message
				fmt.Println("you have to confirm it")
				return nil
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
				cmdData.RestApiClient,
				[]uxHelpers.Process{{
					Id:                  processId,
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
