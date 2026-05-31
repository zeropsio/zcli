package entity

import "github.com/zeropsio/zcli/src/optional"

type Project struct {
	Id          ProjectId
	Name        string
	Mode        ProjectMode
	OrgId       OrgId
	OrgName     string
	Description string
	Status      ProjectStatus
}

var ProjectFields = entityTemplateFields[Project]()

type PostProject struct {
	OrgId        OrgId
	Name         string
	Tags         []string
	Mode         ProjectMode
	SshIsolation string
	EnvIsolation string
	Location     optional.Null[LocationId]
}
