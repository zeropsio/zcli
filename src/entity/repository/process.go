package repository

import (
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func ProcessFromApiOutput(process output.Process) entity.Process {
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
		AppVersion: processAppVersionFromApiOutput(process.AppVersion),
	}
}

func processAppVersionFromApiOutput(appVersion *output.AppVersionJsonObject) *entity.ProcessAppVersion {
	if appVersion == nil {
		return nil
	}
	result := &entity.ProcessAppVersion{
		Id: entity.AppVersionId(appVersion.Id),
	}
	if appVersion.Build != nil {
		result.Build = &entity.ProcessAppVersionStage{ServiceId: serviceIdNullFromOutput(appVersion.Build.ServiceStackId)}
	}
	if appVersion.PrepareCustomRuntime != nil {
		result.PrepareCustomRuntime = &entity.ProcessAppVersionStage{ServiceId: serviceIdNullFromOutput(appVersion.PrepareCustomRuntime.ServiceStackId)}
	}
	return result
}

func serviceIdNullFromOutput(in uuid.ServiceStackIdNull) optional.Null[entity.ServiceId] {
	if sid, ok := in.Get(); ok {
		return optional.New(entity.ServiceId(sid))
	}
	return optional.Empty[entity.ServiceId]()
}
