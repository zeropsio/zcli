package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetServiceByIdOrName(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId uuid.ProjectId,
	serviceIdOrName string,
) (entity.Service, error) {
	service, err := GetServiceById(ctx, restApiClient, uuid.ServiceStackId(serviceIdOrName))
	if err != nil {
		if errorsx.Is(err, errorsx.Or(
			errorsx.ErrorCode(errorCode.InvalidUserInput),
			errorsx.ErrorCode(errorCode.ServiceStackNotFound),
		)) {
			service, err = GetServiceByName(ctx, restApiClient, projectId, types.String(serviceIdOrName))
			if err != nil {
				return entity.Service{}, errorsx.Convert(
					err,
					errorsx.ErrorCode(errorCode.ServiceStackNotFound, errorsx.ErrorCodeErrorMessage(
						func(_ apiError.Error) string {
							return i18n.T(i18n.ErrorServiceNotFound, serviceIdOrName)
						},
					)),
				)
			}
		}
	}

	return service, nil
}

func GetServiceById(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	serviceId uuid.ServiceStackId,
) (entity.Service, error) {
	serviceResponse, err := restApiClient.GetServiceStack(ctx, path.ServiceStackId{Id: serviceId})
	if err != nil {
		return entity.Service{}, err
	}

	serviceOutput, err := serviceResponse.Output()
	if err != nil {
		return entity.Service{}, err
	}

	service := serviceFromApiOutput(serviceOutput)
	return service, nil
}

func GetServiceByName(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId uuid.ProjectId,
	serviceName types.String,
) (entity.Service, error) {
	serviceResponse, err := restApiClient.GetServiceStackByName(ctx, path.GetServiceStackByName{
		ProjectId: projectId,
		Name:      serviceName,
	})
	if err != nil {
		return entity.Service{}, err
	}

	serviceOutput, err := serviceResponse.Output()
	if err != nil {
		return entity.Service{}, err
	}

	service := serviceFromApiOutput(serviceOutput)
	return service, nil
}

func GetNonSystemServicesByProject(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	project entity.Project,
) ([]entity.Service, error) {
	esFilter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    project.Id.TypedString(),
			},
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    project.OrgId.TypedString(),
			},
		},
	}

	servicesResponse, err := restApiClient.PostServiceStackSearch(ctx, esFilter)
	if err != nil {
		return nil, err
	}

	servicesOutput, err := servicesResponse.Output()
	if err != nil {
		return nil, err
	}

	services := make([]entity.Service, 0, len(servicesOutput.Items))
	for _, service := range servicesOutput.Items {
		if !service.IsSystem {
			services = append(services, serviceFromEsSearch(service))
		}
	}

	return services, nil
}

func PostGenericService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	post entity.PostService,
) (entity.Process, entity.Service, error) {
	postBody := body.PostStandardServiceStack{
		Name:             post.Name,
		Mode:             &post.Mode,
		UserDataEnvFile:  post.EnvFile,
		StartWithoutCode: types.NewBoolNull(post.StartWithoutCode.Native()),
		EnvIsolation:     post.EnvIsolation,
		SshIsolation:     post.SshIsolation,
	}
	response, err := restApiClient.PostProjectServiceStack(
		ctx,
		path.ServiceStackServiceStackTypeVersionId{
			Id:                        post.ProjectId,
			ServiceStackTypeVersionId: "runtime",
		},
		postBody,
	)
	if err != nil {
		return entity.Process{}, entity.Service{}, err
	}

	serviceStackProcess, err := response.Output()
	if err != nil {
		return entity.Process{}, entity.Service{}, err
	}

	return processFromApiOutput(serviceStackProcess.Process), serviceFromApiPostOutput(serviceStackProcess), nil
}
func serviceFromEsSearch(esServiceStack output.EsServiceStack) entity.Service {
	return entity.Service{
		Id:                          esServiceStack.Id,
		ProjectId:                   esServiceStack.ProjectId,
		OrgId:                       esServiceStack.ClientId,
		Name:                        esServiceStack.Name,
		Status:                      esServiceStack.Status,
		ServiceTypeId:               esServiceStack.ServiceStackTypeId,
		ServiceTypeCategory:         esServiceStack.ServiceStackTypeInfo.ServiceStackTypeCategory,
		ServiceStackTypeVersionName: esServiceStack.ServiceStackTypeInfo.ServiceStackTypeVersionName,
	}
}

func serviceFromApiOutput(service output.ServiceStack) entity.Service {
	return entity.Service{
		Id:                          service.Id,
		ProjectId:                   service.ProjectId,
		OrgId:                       service.Project.ClientId,
		Name:                        service.Name,
		Status:                      service.Status,
		ServiceTypeId:               service.ServiceStackTypeId,
		ServiceTypeCategory:         service.ServiceStackTypeInfo.ServiceStackTypeCategory,
		ServiceStackTypeVersionName: service.ServiceStackTypeInfo.ServiceStackTypeVersionName,
	}
}

func serviceFromApiPostOutput(service output.ServiceStackProcess) entity.Service {
	return entity.Service{
		Id:                          service.Id,
		ProjectId:                   service.ProjectId,
		OrgId:                       service.Project.ClientId,
		Name:                        service.Name,
		Status:                      service.Status,
		ServiceTypeId:               service.ServiceStackTypeId,
		ServiceTypeCategory:         service.ServiceStackTypeInfo.ServiceStackTypeCategory,
		ServiceStackTypeVersionName: service.ServiceStackTypeInfo.ServiceStackTypeVersionName,
	}
}
