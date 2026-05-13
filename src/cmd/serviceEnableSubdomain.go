package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceEnableSubdomainCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("enable-subdomain").
		Short(i18n.T(i18n.CmdDescServiceEnableSubdomain)).
		ScopeLevel(cmdBuilder.ScopeService()).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpServiceEnableSubdomain)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			enableSubdomainResponse, err := cmdData.RestApiClient.PutServiceStackEnableSubdomainAccess(
				ctx,
				path.ServiceStackId{
					Id: service.Id,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := enableSubdomainResponse.Output()
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
					Running: i18n.T(i18n.ServiceEnablingSubdomain),
					Error:   i18n.T(i18n.ServiceEnableSubdomainFailed),
					Success: i18n.T(i18n.ServiceEnabledSubdomain),
				},
			)
		})
}
