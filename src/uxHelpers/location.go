package uxHelpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

func PrintLocationSelector(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
) (entity.Location, error) {
	locations, err := repository.GetAllLocations(ctx, restApiClient)
	if err != nil {
		return entity.Location{}, err
	}

	if len(locations) == 0 {
		return entity.Location{}, errors.New("No locations available.")
	}

	if len(locations) == 1 {
		return locations[0], nil
	}

	header, body := createLocationTableRows(locations)

	selected, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel("Please, select a location"),
			selector.WithHeader(header),
			selector.WithSetEnableFiltering(true),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return entity.Location{}, err
	}

	if selected > len(locations)-1 {
		return entity.Location{}, errors.New("We couldn't find a location with the index you entered.")
	}

	return locations[selected], nil
}

func createLocationTableRows(locations []entity.Location) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("ID", "Name")
	body := table.NewBody()
	for _, loc := range locations {
		body.AddRow(table.NewRowFromStrings(
			loc.Id.Native(),
			loc.Name.String(),
		))
	}
	return header, body
}
