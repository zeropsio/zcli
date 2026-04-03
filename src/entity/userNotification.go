package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type UserNotification struct {
	Id             uuid.UserNotificationId
	OrgId          uuid.ClientId
	ProjectId      uuid.ProjectIdNull
	ProjectName    types.StringNull
	Type           enum.UserNotificationTypeEnum
	ActionName     types.String
	ActionCreated  types.DateTime
	ActionFinished types.DateTimeNull
	Acknowledged   types.Bool
	CreatedByUser  string
	ServiceNames   []string
	ErrorMessage   types.StringNull
}

var UserNotificationFields = entityTemplateFields[UserNotification]()
