package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Process struct {
	Id         uuid.ProcessId
	OrgId      uuid.ClientId
	ProjectId  uuid.ProjectId
	ServiceId  uuid.ServiceStackIdNull
	ActionName types.String
	Status     enum.ProcessStatusEnum
	Sequence   types.Int
}
