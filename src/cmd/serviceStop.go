package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceStopCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("stop").
		Short(i18n.T(i18n.CmdDescServiceStop)).
		ScopeLevel(scope.Service).
		Arg(scope.ServiceArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpServiceStop)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			stopServiceResponse, err := cmdData.RestApiClient.PutServiceStackStop(
				ctx,
				path.ServiceStackId{
					Id: cmdData.Service.ID,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := stopServiceResponse.Output()
			if err != nil {
				return err
			}

			processId := responseOutput.Id

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ServiceStopping),
					ErrorMessageMessage: i18n.T(i18n.ServiceStopFailed),
					SuccessMessage:      i18n.T(i18n.ServiceStopped),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
