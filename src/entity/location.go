package entity

import (
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/stringId"
)

type Location struct {
	Id   stringId.LocationId
	Name types.String
}
