package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Service struct {
	ID                  uuid.ServiceStackId
	ClientId            uuid.ClientId
	Name                types.String
	Status              enum.ServiceStackStatusEnum
	ServiceTypeId       stringId.ServiceStackTypeId
	ServiceTypeCategory enum.ServiceStackTypeCategoryEnum
}
