package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"
)

func serviceStartCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("start").
		Short(i18n.T(i18n.CmdServiceStart)).
		ScopeLevel(cmdBuilder.Service).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg(), cmdBuilder.OptionalArgLabel("{serviceName | serviceId}")).
		Short(i18n.T(i18n.ServiceStartHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			startServiceResponse, err := cmdData.RestApiClient.PutServiceStackStart(
				ctx,
				path.ServiceStackId{
					Id: cmdData.Service.ID,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := startServiceResponse.Output()
			if err != nil {
				return err
			}

			processId := responseOutput.Id

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ServiceStarting),
					ErrorMessageMessage: i18n.T(i18n.ServiceStarting),
					SuccessMessage:      i18n.T(i18n.ServiceStarted),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
