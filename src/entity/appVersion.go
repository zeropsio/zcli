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

// AppVersionUpload is the result of creating an app version: the id to deploy
// and the URL the source package must be uploaded to.
type AppVersionUpload struct {
	Id        AppVersionId
	UploadUrl string
}
