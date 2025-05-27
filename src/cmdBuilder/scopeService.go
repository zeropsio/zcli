package cmdBuilder

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	dtoPath "github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type ServiceOption gn.Option[serviceScope]

// WithCreateNewService allows 'create new service' option in selector
func WithCreateNewService() ServiceOption {
	return func(s *serviceScope) {
		s.createNew = true
	}
}

func WithProjectScopeOptions(opts ...ProjectOption) ServiceOption {
	return func(s *serviceScope) {
		s.parent = ScopeProject(opts...)
	}
}

type serviceScope struct {
	parent ScopeLevel

	createNew bool
}

func ScopeService(opts ...ServiceOption) ScopeLevel {
	return gn.ApplyOptionsWithDefault(
		serviceScope{
			parent: ScopeProject(),
		},
		opts...,
	)
}

const ServiceArgName = "service-id-or-name"
const serviceFlagName = "service-id"

func (s *serviceScope) AddCommandFlags(cmd *Cmd) {
	cmd.StringFlag(serviceFlagName, "", i18n.T(i18n.ServiceIdFlag), ShortHand("S"))
	s.parent.AddCommandFlags(cmd)
}

func (s *serviceScope) LoadSelectedScope(ctx context.Context, cmd *Cmd, cmdData *LoggedUserCmdData) error {
	var service entity.Service
	var err error

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
		cmdData.Service = optional.New(service)
	}

	// we have to load project, because we need projectId
	if err := s.parent.LoadSelectedScope(ctx, cmd, cmdData); err != nil {
		return err
	}
	project := cmdData.Project.Some()

	if serviceIdOrName, exists := cmdData.Args[ServiceArgName]; exists && !cmdData.Service.Filled() {
		service, err = repository.GetServiceByIdOrName(ctx, cmdData.RestApiClient, project.ID, serviceIdOrName[0])
		if err != nil {
			return err
		}
		cmdData.Service = optional.New(service)
	}

	// interactive selector of service
	if !cmdData.Service.Filled() {
		selectedService, err := uxHelpers.PrintServiceSelector(
			ctx,
			cmdData.RestApiClient,
			project,
			uxHelpers.WithCreateNewService(s.createNew),
		)
		if err != nil {
			return err
		}

		if selectedService.Filled() {
			service = selectedService.Some()
		} else if terminal.IsTerminal() {
			service, err = createNewService(ctx, project, cmdData)
			if err != nil {
				return err
			}
		}

		cmdData.Service = optional.New(service)
	}

	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.SelectedService), service.Name.String()))

	if !cmdData.Project.Filled() {
		if err := s.parent.LoadSelectedScope(ctx, cmd, cmdData); err != nil {
			return err
		}
	}

	return nil
}

func createNewService(ctx context.Context, project entity.Project, cmdData *LoggedUserCmdData) (entity.Service, error) {
	configContent, err := yamlReader.ReadZeropsYamlContent(
		cmdData.UxBlocks,
		cmdData.Params.GetString("workingDir"),
		cmdData.Params.GetString("zeropsYamlPath"),
	)
	if err != nil {
		return entity.Service{}, err
	}

	setups, err := yamlReader.ReadZeropsYamlSetups(configContent)
	if err != nil {
		return entity.Service{}, err
	}

	label := styles.NewStringBuilder()
	label.WriteString("Type ")
	label.WriteStyledString(
		styles.SelectStyle().
			Bold(true),
		"service",
	)
	label.WriteString(" name")

	name, err := uxBlock.Run(
		input.NewRoot(
			ctx,
			input.WithLabel(label.String()),
			input.WithHelpPlaceholder(),
			input.WithPlaceholderStyle(styles.HelpStyle()),
			input.WithoutPrompt(),
			input.WithSetSuggestions(setups),
		),
		input.GetValueFunc,
	)
	if err != nil {
		return entity.Service{}, err
	}

	response, err := cmdData.RestApiClient.PostServiceStack(
		ctx,
		dtoPath.ServiceStackServiceStackTypeVersionId{ServiceStackTypeVersionId: "alpine_v3_21"},
		body.PostStandardServiceStack{
			ProjectId: project.ID,
			Name:      types.NewString(name),
		},
	)
	if err != nil {
		return entity.Service{}, err
	}
	serviceStackProcess, err := response.Output()
	if err != nil {
		if apiError.HasErrorCode(err, errorCode.ServiceStackNameUnavailable) {
			service, err := repository.GetServiceByName(ctx, cmdData.RestApiClient, project.ID, types.NewString(name))
			if err != nil {
				return entity.Service{}, err
			}
			return service, nil
		}
		return entity.Service{}, err
	}

	if err := uxHelpers.ProcessCheckWithSpinner(
		ctx,
		cmdData.UxBlocks,
		[]uxHelpers.Process{
			{
				F:                   uxHelpers.CheckZeropsProcess(serviceStackProcess.Process.Id, cmdData.RestApiClient),
				RunningMessage:      "Creating service",
				ErrorMessageMessage: "Service creation failed",
				SuccessMessage:      "Service created",
			},
		},
	); err != nil {
		return entity.Service{}, err
	}
	service, err := repository.GetServiceById(ctx, cmdData.RestApiClient, serviceStackProcess.Id)
	if err != nil {
		return entity.Service{}, err
	}
	return service, nil
}
