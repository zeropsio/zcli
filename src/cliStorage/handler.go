package cliStorage

import (
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Handler struct {
	*storage.Handler[Data]
}

type Data struct {
	Token          string
	RegionData     region.Data
	ScopeProjectId uuid.ProjectIdNull
}
