package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Container struct {
	ID        uuid.ContainerId
	ClientId  uuid.ClientId
	ProjectId uuid.ProjectId
	ServiceId uuid.ServiceStackId
	Status    enum.ContainerStatusEnum
	Number    types.Int
	Name      types.StringNull
	Hostname  types.StringNull
	Created   types.DateTime
}
