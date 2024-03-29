package scope

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type service struct {
	parent cmdBuilder.ScopeLevel
}

func (s *service) GetParent() cmdBuilder.ScopeLevel {
	return s.parent
}

const ServiceArgName = "serviceIdOrName"
const serviceFlagName = "serviceId"

func (s *service) AddCommandFlags(cmd *cmdBuilder.Cmd) {
	cmd.StringFlag(serviceFlagName, "", i18n.T(i18n.ServiceIdFlag))
}

func (s *service) LoadSelectedScope(ctx context.Context, _ *cmdBuilder.Cmd, cmdData *cmdBuilder.LoggedUserCmdData) error {
	infoText := i18n.SelectedService
	var service *entity.Service
	var err error

	if serviceIdOrName, exists := cmdData.Args[ServiceArgName]; exists {
		service, err = repository.GetServiceByIdOrName(ctx, cmdData.RestApiClient, cmdData.Project.ID, serviceIdOrName[0])
		if err != nil {
			return err
		}
	}

	// service id is passed as a flag
	if serviceId := cmdData.Params.GetString(serviceFlagName); serviceId != "" {
		service, err = repository.GetServiceById(
			ctx,
			cmdData.RestApiClient,
			uuid.ServiceStackId(serviceId),
		)
		if err != nil {
			return errorsx.Convert(
				err,
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ apiError.Error, metaItemTyped map[string]interface{}) string {
							return i18n.T(i18n.ErrorInvalidServiceId, serviceId, metaItemTyped["message"])
						},
					),
				),
			)
		}
	}

	// interactive selector of service
	if service == nil {
		service, err = uxHelpers.PrintServiceSelector(ctx, cmdData.UxBlocks, cmdData.RestApiClient, *cmdData.Project)
		if err != nil {
			return err
		}
	}

	cmdData.Service = service
	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(infoText), cmdData.Service.Name.String()))
	return nil
}
