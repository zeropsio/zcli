package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type UserData struct {
	ID          uuid.UserDataId
	ClientId    uuid.ClientId
	ServiceId   uuid.ServiceStackId
	ServiceName types.String
	Key         types.String
	Content     types.Text
}
