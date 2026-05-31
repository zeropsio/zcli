package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetServiceByIdOrName(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
	serviceIdOrName string,
) (entity.Service, error) {
	service, err := GetServiceById(ctx, restApiClient, entity.ServiceId(serviceIdOrName))
	if err != nil {
		if errorsx.Is(err, errorsx.Or(
			errorsx.ErrorCode(errorsx.CodeInvalidUserInput),
			errorsx.ErrorCode(errorsx.CodeServiceStackNotFound),
		)) {
			service, err = GetServiceByName(ctx, restApiClient, projectId, serviceIdOrName)
			if err != nil {
				return entity.Service{}, errorsx.Convert(
					err,
					errorsx.ErrorCode(errorsx.CodeServiceStackNotFound, errorsx.ErrorCodeErrorMessage(
						func(_ errorsx.ApiError) string {
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
	serviceId entity.ServiceId,
) (entity.Service, error) {
	serviceResponse, err := restApiClient.GetServiceStack(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
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
	projectId entity.ProjectId,
	serviceName string,
) (entity.Service, error) {
	serviceResponse, err := restApiClient.GetServiceStackByName(ctx, path.GetServiceStackByName{
		ProjectId: uuid.ProjectId(projectId),
		Name:      types.NewString(serviceName),
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
	var services []entity.Service
	for offset := 0; ; {
		servicesResponse, err := restApiClient.GetProjectServiceStack(ctx, path.ProjectId{Id: uuid.ProjectId(project.Id)}, query.ListProjectServiceStacks{
			Limit:    types.NewIntNull(listPageLimit),
			Offset:   types.NewIntNull(offset),
			IsSystem: types.NewBoolNull(false),
		})
		if err != nil {
			return nil, err
		}

		servicesOutput, err := servicesResponse.Output()
		if err != nil {
			return nil, err
		}

		for _, service := range servicesOutput.List {
			services = append(services, serviceFromApiOutput(service))
		}

		offset += len(servicesOutput.List)
		if len(servicesOutput.List) == 0 || offset >= servicesOutput.Total.Native() {
			break
		}
	}

	return services, nil
}

func PostGenericService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	post entity.PostService,
) (entity.Process, entity.Service, error) {
	envFile := types.TextNull{}
	if v, ok := post.EnvFile.Get(); ok {
		envFile = types.NewTextNull(v)
	}
	location := stringId.LocationIdNull{}
	if loc, ok := post.Location.Get(); ok {
		location = stringId.NewLocationIdNullFromString(loc.String())
	}
	postBody := body.PostStandardServiceStack{
		Name:             types.NewString(post.Name),
		UserDataEnvFile:  envFile,
		StartWithoutCode: types.NewBoolNull(post.StartWithoutCode),
		EnvIsolation:     types.NewStringNull(post.EnvIsolation),
		SshIsolation:     types.NewStringNull(post.SshIsolation),
		Location:         location,
	}
	response, err := restApiClient.PostProjectServiceStackOld(
		ctx,
		path.ServiceStackServiceStackTypeVersionId{
			Id:                        uuid.ProjectId(post.ProjectId),
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

func StartService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	serviceId entity.ServiceId,
) (entity.Process, error) {
	response, err := restApiClient.PutServiceStackStart(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return entity.Process{}, err
	}
	processOutput, err := response.Output()
	if err != nil {
		return entity.Process{}, err
	}
	return processFromApiOutput(processOutput), nil
}

func StopService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	serviceId entity.ServiceId,
) (entity.Process, error) {
	response, err := restApiClient.PutServiceStackStop(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return entity.Process{}, err
	}
	processOutput, err := response.Output()
	if err != nil {
		return entity.Process{}, err
	}
	return processFromApiOutput(processOutput), nil
}

func DeleteService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	serviceId entity.ServiceId,
) (entity.Process, error) {
	response, err := restApiClient.DeleteServiceStack(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return entity.Process{}, err
	}
	processOutput, err := response.Output()
	if err != nil {
		return entity.Process{}, err
	}
	return processFromApiOutput(processOutput), nil
}

func EnableServiceSubdomainAccess(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	serviceId entity.ServiceId,
) (entity.Process, error) {
	response, err := restApiClient.PutServiceStackEnableSubdomainAccess(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return entity.Process{}, err
	}
	processOutput, err := response.Output()
	if err != nil {
		return entity.Process{}, err
	}
	return processFromApiOutput(processOutput), nil
}

func serviceFromApiOutput(service output.ServiceStack) entity.Service {
	return entity.Service{
		Id:                          entity.ServiceId(service.Id),
		ProjectId:                   entity.ProjectId(service.ProjectId),
		OrgId:                       entity.OrgId(service.Project.ClientId),
		Name:                        service.Name.Native(),
		Status:                      entity.ServiceStatus(service.Status),
		ServiceTypeId:               entity.ServiceTypeId(service.ServiceStackTypeId),
		ServiceTypeCategory:         entity.ServiceTypeCategory(service.ServiceStackTypeInfo.ServiceStackTypeCategory),
		ServiceStackTypeVersionName: service.ServiceStackTypeInfo.ServiceStackTypeVersionName.Native(),
	}
}

func serviceFromApiPostOutput(service output.ServiceStackProcess) entity.Service {
	return entity.Service{
		Id:                          entity.ServiceId(service.Id),
		ProjectId:                   entity.ProjectId(service.ProjectId),
		OrgId:                       entity.OrgId(service.Project.ClientId),
		Name:                        service.Name.Native(),
		Status:                      entity.ServiceStatus(service.Status),
		ServiceTypeId:               entity.ServiceTypeId(service.ServiceStackTypeId),
		ServiceTypeCategory:         entity.ServiceTypeCategory(service.ServiceStackTypeInfo.ServiceStackTypeCategory),
		ServiceStackTypeVersionName: service.ServiceStackTypeInfo.ServiceStackTypeVersionName.Native(),
	}
}
