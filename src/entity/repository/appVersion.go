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

	var searchData []body.EsSearchItem
	searchData = append(searchData, body.EsSearchItem{
		Name:     "clientId",
		Operator: "eq",
		Value:    service.ClientId.TypedString(),
	}, body.EsSearchItem{
		Name:     "serviceStackId",
		Operator: "eq",
		Value:    service.ID.TypedString(),
	}, body.EsSearchItem{
		Name:     "build.serviceStackId",
		Operator: "ne",
		Value:    "",
	})

	response, err := restApiClient.PostAppVersionSearch(ctx, body.EsFilter{
		Search: searchData,
	})
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
	var searchData []body.EsSearchItem
	searchData = append(searchData, body.EsSearchItem{
		Name:     "clientId",
		Operator: "eq",
		Value:    service.ClientId.TypedString(),
	}, body.EsSearchItem{
		Name:     "serviceStackId",
		Operator: "eq",
		Value:    service.ID.TypedString(),
	}, body.EsSearchItem{
		Name:     "build.serviceStackId",
		Operator: "ne",
		Value:    "",
	})
	var sortData []body.EsSortItem
	sortData = append(sortData, body.EsSortItem{
		Name:      "sequence",
		Ascending: types.NewBoolNull(false),
	})

	response, err := restApiClient.PostAppVersionSearch(ctx, body.EsFilter{
		Search: searchData,
		Sort:   sortData,
		Limit:  types.NewIntNull(1),
	})
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
		ClientId:   esAppVersion.ClientId,
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
