package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

func GetAllLocations(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
) ([]entity.Location, error) {
	response, err := restApiClient.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}

	locations := make([]entity.Location, 0, len(resOutput.LocationList))
	for _, loc := range resOutput.LocationList {
		locations = append(locations, entity.Location{
			Id:   loc.Id,
			Name: loc.Name,
		})
	}

	return locations, nil
}
