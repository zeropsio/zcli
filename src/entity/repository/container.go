package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
)

func GetAllContainers(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.Container, error) {
	var searchData []body.EsSearchItem
	searchData = append(searchData, body.EsSearchItem{
		Name:     "clientId",
		Operator: "eq",
		Value:    service.ClientId.TypedString(),
	}, body.EsSearchItem{
		Name:     "serviceStackId",
		Operator: "eq",
		Value:    service.ID.TypedString(),
	})

	response, err := restApiClient.PostContainerSearch(ctx, body.EsFilter{
		Search: searchData,
	})
	if err != nil {
		return nil, err
	}

	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}

	containers := make([]entity.Container, 0, len(resOutput.Items))
	for _, container := range resOutput.Items {
		containers = append(containers, containerFromEsSearch(container))
	}

	return containers, nil
}

func containerFromEsSearch(esContainer output.EsContainer) entity.Container {

	return entity.Container{
		ID:        esContainer.Id,
		ClientId:  esContainer.ClientId,
		ProjectId: esContainer.ProjectId,
		ServiceId: esContainer.ServiceStackId,
		Status:    esContainer.Status,
		Number:    esContainer.Number,
		Name:      esContainer.Name,
		Hostname:  esContainer.Hostname,
		Created:   esContainer.Created,
	}
}
