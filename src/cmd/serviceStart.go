package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func serviceStartCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("start").
		Short(i18n.T(i18n.CmdDescServiceStart)).
		ScopeLevel(cmdBuilder.ScopeService()).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg(), cmdBuilder.OptionalArgLabel("{service-name | service-id}")).
		HelpFlag(i18n.T(i18n.CmdHelpServiceStart)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			process, err := repository.StartService(ctx, cmdData.RestApiClient, service.Id)
			if err != nil {
				return err
			}

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(process.Id, cmdData.RestApiClient, cmdData.UxBlocks),
					RunningMessage:      i18n.T(i18n.ServiceStarting),
					ErrorMessageMessage: i18n.T(i18n.ServiceStartFailed),
					SuccessMessage:      i18n.T(i18n.ServiceStarted),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
