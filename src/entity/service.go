package entity

import (
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

type PostService struct {
	ProjectId        uuid.ProjectId
	Name             types.String
	Mode             enum.ServiceStackModeEnum
	EnvFile          types.TextNull
	StartWithoutCode types.Bool
	SshIsolation     types.StringNull
	EnvIsolation     types.StringNull
}
