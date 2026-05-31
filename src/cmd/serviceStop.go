package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/uxHelpers"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceStopCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("stop").
		Short(i18n.T(i18n.CmdDescServiceStop)).
		ScopeLevel(cmdBuilder.ScopeService()).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpServiceStop)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			process, err := repository.StopService(ctx, cmdData.RestApiClient, service.Id)
			if err != nil {
				return err
			}

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(process.Id, cmdData.RestApiClient, cmdData.UxBlocks),
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
