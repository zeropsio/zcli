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
}
