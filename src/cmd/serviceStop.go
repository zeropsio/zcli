package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceStopCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("stop").
		Short(i18n.T(i18n.CmdServiceStop)).
		ScopeLevel(cmdBuilder.Service).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		Short(i18n.T(i18n.ServiceStopHelp)).
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
					ErrorMessageMessage: i18n.T(i18n.ServiceStopping),
					SuccessMessage:      i18n.T(i18n.ServiceStopped),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
