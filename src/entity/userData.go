package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type UserData struct {
	ID             uuid.UserDataId
	ClientId       uuid.ClientId
	ServiceStackId uuid.ServiceStackId
	Key            types.String
	Content        types.Text
}
