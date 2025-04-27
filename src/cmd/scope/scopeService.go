package scope

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	zeropsUuid "github.com/zeropsio/zcli/src/uuid"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type service struct {
	parent cmdBuilder.ScopeLevel
}

const ServiceArgName = "serviceId"
const serviceFlagName = "serviceId"

func (s *service) AddCommandFlags(cmd *cmdBuilder.Cmd) {
	cmd.StringFlag(
		serviceFlagName, 
		"", 
		i18n.T(i18n.ServiceIdFlag), 
		cmdBuilder.DeprecatedFlag("Use positional parameter instead: zcli command [serviceId]"),
	)
	s.parent.AddCommandFlags(cmd)
}

func (s *service) LoadSelectedScope(ctx context.Context, cmd *cmdBuilder.Cmd, cmdData *cmdBuilder.LoggedUserCmdData) error {
	infoText := i18n.SelectedService
	var service *entity.Service
	var err error

	// First check for positional argument (serviceId)
	if serviceIdValues, exists := cmdData.Args[ServiceArgName]; exists && service == nil {
		serviceIdStr := serviceIdValues[0]
		// Only accept valid UUIDs as positional parameters
		if zeropsUuid.IsValidServiceId(serviceIdStr) {
			service, err = repository.GetServiceById(
				ctx,
				cmdData.RestApiClient,
				uuid.ServiceStackId(serviceIdStr),
			)
			if err != nil {
				return errorsx.Convert(
					err,
					errorsx.InvalidUserInput(
						"id",
						errorsx.InvalidUserInputErrorMessage(
							func(_ apiError.Error, metaItemTyped map[string]interface{}) string {
								return i18n.T(i18n.ErrorInvalidServiceId, serviceIdStr, metaItemTyped["message"])
							},
						),
					),
				)
			}
		} else {
			return errorsx.Convert(
				fmt.Errorf("invalid service ID format"),
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ apiError.Error, _ map[string]interface{}) string {
							return fmt.Sprintf("Invalid service ID format: '%s'. Please provide a valid service ID UUID.", serviceIdStr)
						},
					),
				),
			)
		}
	}

	// Fall back to service id flag if positional argument is not provided
	if service == nil && cmdData.Params.GetString(serviceFlagName) != "" {
		serviceId := cmdData.Params.GetString(serviceFlagName)
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
		// we have to load project, because we need projectId
		if err := s.parent.LoadSelectedScope(ctx, cmd, cmdData); err != nil {
			return err
		}
		service, err = uxHelpers.PrintServiceSelector(ctx, cmdData.UxBlocks, cmdData.RestApiClient, *cmdData.Project)
		if err != nil {
			return err
		}
	}

	cmdData.Service = service
	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(infoText), cmdData.Service.Name.String()))

	// Load parent scope from selected service if it wasn't loaded already above
	if cmdData.Project == nil {
		if err := s.parent.LoadSelectedScope(ctx, cmd, cmdData); err != nil {
			return err
		}
	}

	return nil
}