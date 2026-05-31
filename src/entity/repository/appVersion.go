package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetLatestAppVersionByService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.AppVersion, error) {
	response, err := restApiClient.GetServiceStackAppVersion(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(service.Id)}, query.ListServiceStackAppVersions{
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
		Id:         entity.AppVersionId(appVersion.Id),
		OrgId:      entity.OrgId(appVersion.ClientId),
		ProjectId:  entity.ProjectId(appVersion.ProjectId),
		ServiceId:  entity.ServiceId(appVersion.ServiceStackId),
		Source:     entity.AppVersionSource(appVersion.Source),
		Sequence:   int(appVersion.Sequence),
		Status:     entity.AppVersionStatus(appVersion.Status),
		Created:    appVersion.Created.Native(),
		LastUpdate: appVersion.LastUpdate.Native(),
		Build:      appVersionBuildFromApiOutput(appVersion.Build),
	}
}

func appVersionBuildFromApiOutput(build *output.AppVersionBuild) *entity.AppVersionBuild {
	if build == nil {
		return nil
	}
	serviceStackId := optional.Empty[entity.ServiceId]()
	if sid, ok := build.ServiceStackId.Get(); ok {
		serviceStackId = optional.New(entity.ServiceId(sid))
	}
	name, _ := build.ServiceStackName.Get()
	return &entity.AppVersionBuild{
		ServiceStackId:   serviceStackId,
		ServiceStackName: name.Native(),
	}
}
