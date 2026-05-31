package entity

import "github.com/zeropsio/zcli/src/optional"

type Service struct {
	Id                          ServiceId
	ProjectId                   ProjectId
	OrgId                       OrgId
	Name                        string
	Status                      ServiceStatus
	ServiceTypeId               ServiceTypeId
	ServiceTypeCategory         ServiceTypeCategory
	ServiceStackTypeVersionName string
}

var ServiceFields = entityTemplateFields[Service]()

type PostService struct {
	ProjectId        ProjectId
	Name             string
	EnvFile          optional.Null[string]
	StartWithoutCode bool
	SshIsolation     string
	EnvIsolation     string
	Location         optional.Null[LocationId]
}
