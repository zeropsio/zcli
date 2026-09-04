package entity

import (
	"slices"

	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Service struct {
	Id                          uuid.ServiceStackId
	ProjectId                   uuid.ProjectId
	OrgId                       uuid.ClientId
	Name                        types.String
	Status                      enum.ServiceStackStatusEnum
	ServiceTypeId               stringId.ServiceStackTypeId
	ServiceTypeCategory         enum.ServiceStackTypeCategoryEnum
	ServiceStackTypeVersionName types.String
}

var ServiceFields = entityTemplateFields[Service]()

// ServiceMode mirrors the ServiceStackModeEnum the SDK dropped when the platform deprecated serviceStack.mode to a plain string.
type ServiceMode string

const (
	ServiceModeHa    ServiceMode = "HA"
	ServiceModeNonHa ServiceMode = "NON_HA"
)

func ServiceModeAll() []ServiceMode {
	return []ServiceMode{ServiceModeHa, ServiceModeNonHa}
}

func (m ServiceMode) Is(values ...ServiceMode) bool {
	return slices.Contains(values, m)
}

type PostService struct {
	ProjectId        uuid.ProjectId
	Name             types.String
	Mode             ServiceMode
	EnvFile          types.TextNull
	StartWithoutCode types.Bool
	SshIsolation     types.StringNull
	EnvIsolation     types.StringNull
	Location         stringId.LocationIdNull
}
