package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func ProjectImport(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgId entity.OrgId,
	yaml string,
) ([]entity.ImportServiceStack, error) {
	response, err := restApiClient.PostClientProjectImport(
		ctx,
		path.ClientId{Id: uuid.ClientId(orgId)},
		body.ProjectImport{Yaml: types.Text(yaml)},
	)
	if err != nil {
		return nil, err
	}
	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}
	return importServiceStacksFromOutput(resOutput.ServiceStacks), nil
}

func ServiceImport(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
	yaml string,
) ([]entity.ImportServiceStack, error) {
	response, err := restApiClient.PostProjectServiceStackImport(
		ctx,
		path.ProjectId{Id: uuid.ProjectId(projectId)},
		body.ServiceStackImport{Yaml: types.Text(yaml)},
	)
	if err != nil {
		return nil, err
	}
	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}
	return importServiceStacksFromOutput(resOutput.ServiceStacks), nil
}

func importServiceStacksFromOutput(in output.ProjectImportServiceStacks) []entity.ImportServiceStack {
	result := make([]entity.ImportServiceStack, 0, len(in))
	for _, s := range in {
		result = append(result, entity.ImportServiceStack{
			Id:        entity.ServiceId(s.Id),
			Name:      s.Name.Native(),
			Processes: gn.TransformSlice(s.Processes, ProcessFromApiOutput),
		})
	}
	return result
}
