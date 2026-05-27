package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
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
	response, err := restApiClient.GetProjectProcess(
		ctx,
		path.ProjectId{Id: projectId},
		query.ListProjectProcesses{
			ActionNameContains: actionName.StringNull(),
		},
	)
	if err != nil {
		return nil, err
	}
	processList, err := response.Output()
	if err != nil {
		return nil, err
	}
	return gn.TransformSlice(processList.List, processFromApiOutput), nil
}

func processFromApiOutput(process output.Process) entity.Process {
	return entity.Process{
		Id:         process.Id,
		OrgId:      process.ClientId,
		ProjectId:  process.ProjectId,
		ServiceId:  process.ServiceStackId,
		ActionName: process.ActionName,
		Status:     process.Status,
		Sequence:   process.Sequence,
	}
}
