package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceEnableSubdomainCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("enable-subdomain").
		Short(i18n.T(i18n.CmdServiceEnableSubdomain)).
		ScopeLevel(scope.Service).
		Arg(scope.ServiceArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.ServiceEnableSubdomainHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			enableSubdomainResponse, err := cmdData.RestApiClient.PutServiceStackEnableSubdomainAccess(
				ctx,
				path.ServiceStackId{
					Id: cmdData.Service.ID,
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

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ServiceEnablingSubdomain),
					ErrorMessageMessage: i18n.T(i18n.ServiceEnablingSubdomain),
					SuccessMessage:      i18n.T(i18n.ServiceEnabledSubdomain),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
