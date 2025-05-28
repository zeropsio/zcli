package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
)

func GetAllAppVersionByService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.AppVersion, error) {
	esFilter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    service.OrgId.TypedString(),
			}, {
				Name:     "serviceStackId",
				Operator: "eq",
				Value:    service.Id.TypedString(),
			}, {
				Name:     "build.serviceStackId",
				Operator: "ne",
				Value:    "",
			},
		},
	}

	response, err := restApiClient.PostAppVersionSearch(ctx, esFilter)
	if err != nil {
		return nil, err
	}

	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}

	appVersions := make([]entity.AppVersion, 0, len(resOutput.Items))
	for _, appVersion := range resOutput.Items {
		appVersions = append(appVersions, appVersionFromEsSearch(appVersion))
	}

	return appVersions, nil
}

func GetLatestAppVersionByService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.AppVersion, error) {
	esFilter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    service.OrgId.TypedString(),
			}, {
				Name:     "serviceStackId",
				Operator: "eq",
				Value:    service.Id.TypedString(),
			}, {
				Name:     "build.serviceStackId",
				Operator: "ne",
				Value:    "",
			},
		},
		Sort: []body.EsSortItem{
			{
				Name:      "sequence",
				Ascending: types.NewBoolNull(false),
			},
		},
		Limit: types.NewIntNull(1),
	}

	response, err := restApiClient.PostAppVersionSearch(ctx, esFilter)
	if err != nil {
		return nil, err
	}
	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}

	appVersions := make([]entity.AppVersion, 0, len(resOutput.Items))
	for _, appVersion := range resOutput.Items {
		appVersions = append(appVersions, appVersionFromEsSearch(appVersion))
	}

	return appVersions, nil
}

func appVersionFromEsSearch(esAppVersion output.EsAppVersion) entity.AppVersion {
	return entity.AppVersion{
		Id:         esAppVersion.Id,
		OrgId:      esAppVersion.ClientId,
		ProjectId:  esAppVersion.ProjectId,
		ServiceId:  esAppVersion.ServiceStackId,
		Source:     esAppVersion.Source,
		Sequence:   esAppVersion.Sequence,
		Status:     esAppVersion.Status,
		Created:    esAppVersion.Created,
		LastUpdate: esAppVersion.LastUpdate,
		Build:      esAppVersion.Build,
	}
}
