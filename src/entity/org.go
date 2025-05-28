package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Org struct {
	Id   uuid.ClientId
	Role enum.ClientUserRoleCodeEnum
	Name types.String
}

var OrgFields = entityTemplateFields[Org]()
