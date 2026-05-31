package entity

import "github.com/zeropsio/zcli/src/optional"

type Process struct {
	Id         ProcessId
	OrgId      OrgId
	ProjectId  ProjectId
	ServiceId  optional.Null[ServiceId]
	ActionName string
	Status     ProcessStatus
	Sequence   int
	// AppVersion is set for build/deploy processes and drives the push/deploy
	// log streaming. It is nil for processes that don't carry an app version.
	AppVersion *ProcessAppVersion
}

type ProcessAppVersion struct {
	Id AppVersionId
	// Build is non-nil once the build stage exists; ServiceId points at the build service.
	Build *ProcessAppVersionStage
	// PrepareCustomRuntime is non-nil once the prepare-runtime stage exists.
	PrepareCustomRuntime *ProcessAppVersionStage
}

type ProcessAppVersionStage struct {
	ServiceId optional.Null[ServiceId]
}
