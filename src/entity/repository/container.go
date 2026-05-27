package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
)

func GetAllContainers(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
) ([]entity.Container, error) {
	response, err := restApiClient.GetServiceStackContainer(
		ctx,
		path.ServiceStackId{Id: service.Id},
		query.ListServiceStackContainers{},
	)
	if err != nil {
		return nil, err
	}

	containerList, err := response.Output()
	if err != nil {
		return nil, err
	}

	return gn.TransformSlice(containerList.List, containerFromListOutput), nil
}

func containerFromListOutput(container output.Container) entity.Container {
	return entity.Container{
		Id:        container.Id,
		OrgId:     container.ClientId,
		ProjectId: container.ProjectId,
		ServiceId: container.ServiceStackId,
		Status:    gn.Ptr(container.Status),
		Number:    container.Number,
		Name:      container.Name,
		Hostname:  container.Hostname,
		Created:   container.Created,
	}
}
