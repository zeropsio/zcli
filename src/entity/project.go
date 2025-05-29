package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Project struct {
	Id          uuid.ProjectId
	Name        types.String
	Mode        enum.ProjectModeEnum
	OrgId       uuid.ClientId
	OrgName     types.String
	Description types.Text
	Status      enum.ProjectStatusEnum
}

var ProjectFields = entityTemplateFields[Project]()

type PostProject struct {
	OrgId        uuid.ClientId
	Name         types.String
	Tags         types.StringArray
	Mode         enum.ProjectModeEnum
	SshIsolation types.StringNull
	EnvIsolation types.StringNull
}
