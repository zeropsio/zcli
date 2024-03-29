package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Project struct {
	ID          uuid.ProjectId
	Name        types.String
	OrgId       uuid.ClientId
	OrgName     types.String
	Description types.Text
	Status      enum.ProjectStatusEnum
}
