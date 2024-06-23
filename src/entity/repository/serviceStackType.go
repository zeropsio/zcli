package repository

import (
	"context"
	"errors"
	"sort"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types/stringId"
)

func GetServiceStackTypes(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
) (result []entity.ServiceStackType, _ error) {
	settings, err := restApiClient.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	settingsOutput, err := settings.Output()
	if err != nil {
		return nil, err
	}
	for _, serviceStackType := range settingsOutput.ServiceStackList {
		e := entity.ServiceStackType{
			ID:   serviceStackType.Id,
			Name: serviceStackType.Name,
		}

		for _, serviceStackTypeVersion := range serviceStackType.ServiceStackTypeVersionList {
			if !serviceStackTypeVersion.Status.IsActive() {
				continue
			}
			if serviceStackTypeVersion.IsBuild.Native() {
				continue
			}
			if serviceStackTypeVersion.Name.Native() == "prepare_runtime" {
				continue
			}
			e.Versions = append(e.Versions, entity.ServiceStackTypeVersion{
				ID:                 serviceStackTypeVersion.Id,
				Name:               serviceStackTypeVersion.Name,
				ExactVersionNumber: serviceStackTypeVersion.ExactVersionNumber,
			})
		}
		if len(e.Versions) == 0 {
			continue
		}
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name.Native() < result[j].Name.Native() })
	return result, nil
}

func GetServiceStackTypeById(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	serviceStackTypeId stringId.ServiceStackTypeId,
) (result entity.ServiceStackType, _ error) {
	serviceStackTypes, err := GetServiceStackTypes(ctx, restApiClient)
	if err != nil {
		return result, err
	}
	for _, serviceStackType := range serviceStackTypes {
		if serviceStackType.ID == serviceStackTypeId {
			return serviceStackType, nil
		}
		if serviceStackType.Name.Native() == serviceStackTypeId.Native() {
			return serviceStackType, nil
		}
		for _, serviceStackTypeVersion := range serviceStackType.Versions {
			if serviceStackTypeVersion.ID.Native() == serviceStackTypeId.Native() {
				return serviceStackType, nil
			}
			if serviceStackTypeVersion.Name.Native() == serviceStackTypeId.Native() {
				return serviceStackType, nil
			}
		}
	}
	return result, errors.New(string(errorCode.ServiceStackTypeVersionNotFound))
}
