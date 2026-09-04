package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
)

func GetLatestAppVersionByService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.AppVersion, error) {
	response, err := restApiClient.GetServiceStackAppVersion(ctx, path.ServiceStackId{Id: service.Id}, query.ListServiceStackAppVersions{
		Limit:    types.NewIntNull(1),
		HasBuild: types.NewBoolNull(true),
		Sort:     types.NewStringArrayNull(types.StringArray{"-sequence"}),
	})
	if err != nil {
		return nil, err
	}
	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}

	appVersions := make([]entity.AppVersion, 0, len(resOutput.List))
	for _, appVersion := range resOutput.List {
		appVersions = append(appVersions, appVersionFromApiOutput(appVersion))
	}

	return appVersions, nil
}

func appVersionFromApiOutput(appVersion output.GetAppVersion) entity.AppVersion {
	return entity.AppVersion{
		Id:         appVersion.Id,
		OrgId:      appVersion.ClientId,
		ProjectId:  appVersion.ProjectId,
		ServiceId:  appVersion.ServiceStackId,
		Source:     appVersion.Source,
		Sequence:   appVersion.Sequence,
		Status:     appVersion.Status,
		Created:    appVersion.Created,
		LastUpdate: appVersion.LastUpdate,
		Build:      appVersion.Build,
	}
}
