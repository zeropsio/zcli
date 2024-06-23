package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/stringId"
)

type ServiceStackType struct {
	ID       stringId.ServiceStackTypeId
	Name     types.String
	Versions []ServiceStackTypeVersion
}

type ServiceStackTypeVersion struct {
	ID                 stringId.ServiceStackTypeVersionId
	Name               types.String
	ExactVersionNumber types.EmptyString
}
