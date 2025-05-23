package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Process struct {
	ID         uuid.ProcessId
	OrgID      uuid.ClientId
	ProjectID  uuid.ProjectId
	ServiceID  uuid.ServiceStackIdNull
	ActionName types.String
	Status     enum.ProcessStatusEnum
	Sequence   types.Int
}
