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
		Short(i18n.T(i18n.CmdDescServiceStart)).
		ScopeLevel(cmdBuilder.ScopeService()).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg(), cmdBuilder.OptionalArgLabel("{service-name | service-id}")).
		HelpFlag(i18n.T(i18n.CmdHelpServiceStart)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			startServiceResponse, err := cmdData.RestApiClient.PutServiceStackStart(
				ctx,
				path.ServiceStackId{
					Id: service.Id,
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

			return uxHelpers.WatchProcess(
				ctx,
				cmdData.UxBlocks,
				cmdData.RestApiClient,
				processId,
				uxHelpers.ProcessMessages{
					Running: i18n.T(i18n.ServiceStarting),
					Error:   i18n.T(i18n.ServiceStartFailed),
					Success: i18n.T(i18n.ServiceStarted),
				},
			)
		})
}
