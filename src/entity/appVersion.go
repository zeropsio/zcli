package entity

import (
	"time"

	"github.com/zeropsio/zcli/src/optional"
)

type AppVersion struct {
	Id         AppVersionId
	OrgId      OrgId
	ProjectId  ProjectId
	ServiceId  ServiceId
	Source     AppVersionSource
	Sequence   int
	Status     AppVersionStatus
	Created    time.Time
	LastUpdate time.Time
	Build      *AppVersionBuild
}

type AppVersionBuild struct {
	ServiceStackId   optional.Null[ServiceId]
	ServiceStackName string
}
