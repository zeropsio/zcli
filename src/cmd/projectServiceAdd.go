package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
)

const serviceAddArgName = "serviceAddName"
const serviceAddArgType = "type"
const serviceAddArgHa = "ha"

func projectServiceAddCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("service-add").
		Short(i18n.T(i18n.CmdDescProjectServiceAdd)).
		ScopeLevel(scope.Project).
		Arg(serviceAddArgName).
		StringFlag(serviceAddArgType, "", i18n.T(i18n.ServiceAddTypeFlag)).
		BoolFlag(serviceAddArgHa, false, i18n.T(i18n.ServiceAddHaFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpProjectServiceAdd)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			name := cmdData.Args[serviceAddArgName][0]

			var typeNameVersion entity.ServiceStackTypeVersion
			var typeNameVersionId stringId.ServiceStackTypeVersionId

			if cmdData.Params.GetString(serviceAddArgType) == "" {
				serviceStackType, err := uxHelpers.PrintServiceStackTypeSelector(ctx, cmdData.UxBlocks, cmdData.RestApiClient)
				if err != nil {
					return err
				}
				if len(serviceStackType.Versions) == 1 {
					typeNameVersion = serviceStackType.Versions[0]
				} else {
					typeNameVersion, err = uxHelpers.PrintServiceStackTypeVersionSelector(ctx, cmdData.UxBlocks, cmdData.RestApiClient,
						uxHelpers.PrintServiceStackTypeVersionSelectorWithServiceStackTypeIdFilter(serviceStackType),
					)
					if err != nil {
						return err
					}
				}
				typeNameVersionId = typeNameVersion.ID
			} else {
				input := cmdData.Params.GetString(serviceAddArgType)
				serviceStackType, err := repository.GetServiceStackTypeById(ctx, cmdData.RestApiClient, stringId.ServiceStackTypeId(input))
				if err != nil {
					return err
				}
				typeNameVersionId = serviceStackType.Versions[0].ID
			}

			mode := enum.ServiceStackModeEnumNonHa
			if cmdData.Params.GetBool(serviceAddArgHa) {
				mode = enum.ServiceStackModeEnumHa
			}

			serviceAddResponse, err := cmdData.RestApiClient.PostServiceStack(ctx,
				path.ServiceStackServiceStackTypeVersionId{
					ServiceStackTypeVersionId: typeNameVersionId,
				},
				body.PostStandardServiceStack{
					ProjectId: cmdData.Project.ID,
					Name:      types.NewString(name),
					Mode:      &mode,
				},
			)
			if err != nil {
				return err
			}

			serviceAddOutput, err := serviceAddResponse.Output()
			if err != nil {
				return err
			}

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(serviceAddOutput.Process.Id, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ServiceAdding),
					ErrorMessageMessage: i18n.T(i18n.ServiceAddFailed),
					SuccessMessage:      i18n.T(i18n.ServiceAdded),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
