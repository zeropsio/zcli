package repository

import (
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zerops-go/dto/output"
)

func processFromApiOutput(process output.Process) entity.Process {
	serviceId := optional.Empty[entity.ServiceId]()
	if sid, ok := process.ServiceStackId.Get(); ok {
		serviceId = optional.New(entity.ServiceId(sid))
	}
	return entity.Process{
		Id:         entity.ProcessId(process.Id),
		OrgId:      entity.OrgId(process.ClientId),
		ProjectId:  entity.ProjectId(process.ProjectId),
		ServiceId:  serviceId,
		ActionName: process.ActionName.Native(),
		Status:     entity.ProcessStatus(process.Status),
		Sequence:   int(process.Sequence),
	}
}
