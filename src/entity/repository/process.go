package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetProcessByActionNameAndProjectId(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgId uuid.ClientId,
	projectId uuid.ProjectId,
	actionName types.String,
) ([]entity.Process, error) {
	search, err := restApiClient.PostProcessSearch(ctx, body.EsFilter{
		Search: body.EsFilterSearch{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    orgId.TypedString(),
			},
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    projectId.TypedString(),
			},
			{
				Name:     "actionName",
				Operator: "eq",
				Value:    actionName,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	response, err := search.Output()
	if err != nil {
		return nil, err
	}
	return gn.TransformSlice(response.Items, processFromEsSearch), nil
}

func processFromEsSearch(esProcess output.EsProcess) entity.Process {
	return entity.Process{
		ID:         esProcess.Id,
		OrgID:      esProcess.ClientId,
		ProjectID:  esProcess.ProjectId,
		ServiceID:  esProcess.ServiceStackId,
		ActionName: esProcess.ActionName,
		Status:     esProcess.Status,
		Sequence:   esProcess.Sequence,
	}
}

func processFromApiOutput(process output.Process) entity.Process {
	return entity.Process{
		ID:         process.Id,
		OrgID:      process.ClientId,
		ProjectID:  process.ProjectId,
		ServiceID:  process.ServiceStackId,
		ActionName: process.ActionName,
		Status:     process.Status,
		Sequence:   process.Sequence,
	}
}
