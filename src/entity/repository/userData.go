package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/options"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetUserDataByServiceIdOrName(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId uuid.ProjectId,
	serviceIdOrName string,
) ([]entity.UserData, error) {
	project, err := GetProjectById(ctx, restApiClient, projectId)
	if err != nil {
		return nil, err
	}
	service, err := GetServiceById(ctx, restApiClient, uuid.ServiceStackId(serviceIdOrName))
	if err != nil {
		if errorsx.Is(err, errorsx.Or(
			errorsx.ErrorCode(errorCode.InvalidUserInput),
			errorsx.ErrorCode(errorCode.ServiceStackNotFound),
		)) {
			service, err = GetServiceByName(ctx, restApiClient, projectId, types.String(serviceIdOrName))
			if err != nil {
				return nil, errorsx.Convert(
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
	return GetUserDataByServiceId(ctx, restApiClient, project, service.ID, GetUserDataSetup{})
}

type EsFilterFunc func(body.EsFilter) body.EsFilter
type EsFilterFuncs []EsFilterFunc

func (fs EsFilterFuncs) apply(esFilter body.EsFilter) body.EsFilter {
	for _, f := range fs {
		esFilter = f(esFilter)
	}
	return esFilter
}

type GetUserDataSetup struct {
	filters EsFilterFuncs
}

func UserDataSetup(
	opts ...options.Option[GetUserDataSetup],
) GetUserDataSetup {
	return options.ApplyOptions(opts...)
}

func WithUserDataEsFilters(filters ...EsFilterFunc) options.Option[GetUserDataSetup] {
	return func(s *GetUserDataSetup) {
		s.filters = append(s.filters, filters...)
	}
}

func (s *GetUserDataSetup) EsFilters(filters ...EsFilterFunc) {
	s.filters = append(s.filters, filters...)
}

func GetUserDataByProjectId(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	project *entity.Project,
	setup GetUserDataSetup,
) ([]entity.UserData, error) {
	esFilter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    project.ID.TypedString(),
			},
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    project.OrgId.TypedString(),
			},
		},
	}

	return getEsSearchUserData(ctx, restApiClient, setup.filters.apply(esFilter))
}

func GetUserDataByServiceId(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	project *entity.Project,
	serviceId uuid.ServiceStackId,
	setup GetUserDataSetup,
) ([]entity.UserData, error) {
	esFilter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    project.ID.TypedString(),
			},
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    project.OrgId.TypedString(),
			},
			{
				Name:     "serviceStackId",
				Operator: "eq",
				Value:    serviceId.TypedString(),
			},
		},
	}

	return getEsSearchUserData(ctx, restApiClient, setup.filters.apply(esFilter))
}

func getEsSearchUserData(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	esFilter body.EsFilter,
) ([]entity.UserData, error) {
	userDataResponse, err := restApiClient.PostUserDataSearch(ctx, esFilter)
	if err != nil {
		return nil, err
	}

	userDataOutput, err := userDataResponse.Output()
	if err != nil {
		return nil, err
	}

	userDataResult := make([]entity.UserData, 0, len(userDataOutput.Items))
	for _, userData := range userDataOutput.Items {
		userDataResult = append(userDataResult, userDataFromEsSearch(userData))
	}

	return userDataResult, nil
}

func userDataFromEsSearch(userData output.EsUserData) entity.UserData {
	return entity.UserData{
		ID:          userData.Id,
		ClientId:    userData.ClientId,
		ServiceId:   userData.ServiceStackId,
		ServiceName: userData.ServiceStackName,
		Key:         userData.Key,
		Content:     userData.Content,
	}
}
