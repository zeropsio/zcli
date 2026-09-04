package repository

import (
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zerops-go/dto/output"
)

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
