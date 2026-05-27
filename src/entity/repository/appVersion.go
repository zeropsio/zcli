package repository

import (
	"context"
	"sort"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
)

func GetAllAppVersionByService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.AppVersion, error) {
	response, err := restApiClient.GetServiceStackAppVersion(
		ctx,
		path.ServiceStackId{Id: service.Id},
		query.ListServiceStackAppVersions{},
	)
	if err != nil {
		return nil, err
	}

	versionList, err := response.Output()
	if err != nil {
		return nil, err
	}

	appVersions := make([]entity.AppVersion, 0, len(versionList.List))
	for _, v := range versionList.List {
		if v.Build != nil {
			appVersions = append(appVersions, appVersionFromListOutput(v))
		}
	}

	return appVersions, nil
}

func GetLatestAppVersionByService(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.AppVersion, error) {
	response, err := restApiClient.GetServiceStackAppVersion(
		ctx,
		path.ServiceStackId{Id: service.Id},
		query.ListServiceStackAppVersions{},
	)
	if err != nil {
		return nil, err
	}

	versionList, err := response.Output()
	if err != nil {
		return nil, err
	}

	appVersions := make([]entity.AppVersion, 0, len(versionList.List))
	for _, v := range versionList.List {
		if v.Build != nil {
			appVersions = append(appVersions, appVersionFromListOutput(v))
		}
	}

	sort.Slice(appVersions, func(i, j int) bool {
		return appVersions[i].Sequence.Native() > appVersions[j].Sequence.Native()
	})

	if len(appVersions) > 1 {
		appVersions = appVersions[:1]
	}

	return appVersions, nil
}

func appVersionFromListOutput(av output.GetAppVersion) entity.AppVersion {
	return entity.AppVersion{
		Id:         av.Id,
		OrgId:      av.ClientId,
		ProjectId:  av.ProjectId,
		ServiceId:  av.ServiceStackId,
		Source:     av.Source,
		Sequence:   av.Sequence,
		Status:     av.Status,
		Created:    av.Created,
		LastUpdate: av.LastUpdate,
		Build:      av.Build,
	}
}
