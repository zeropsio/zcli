package entity

import (
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type AppVersion struct {
	Id         uuid.AppVersionId
	ClientId   uuid.ClientId
	ProjectId  uuid.ProjectId
	ServiceId  uuid.ServiceStackId
	Source     enum.AppVersionSourceEnum
	Sequence   types.Int
	Status     enum.AppVersionStatusEnum
	Created    types.DateTime
	LastUpdate types.DateTime
	Build      *output.AppVersionBuild
}

type AppVersionBuild struct {
	ServiceStackId            uuid.ServiceStackIdNull
	ServiceStackName          types.StringNull
	ServiceStackTypeVersionId stringId.ServiceStackTypeVersionIdNull
	PipelineStart             types.DateTimeNull
	PipelineFinish            types.DateTimeNull
	PipelineFailed            types.DateTimeNull
	ContainerCreationStart    types.DateTimeNull
	StartDate                 types.DateTimeNull
	EndDate                   types.DateTimeNull
	CacheUsed                 types.Bool
	HasCurrentCache           types.Bool
}
