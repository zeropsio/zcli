package cliStorage

import (
	"time"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Handler struct {
	*storage.Handler[Data]
}

type Data struct {
	// Token is a static personal or integration token. External tools write it too, so its meaning must not change.
	Token string
	// LoginSession is set by the browser login. At most one of Token and LoginSession is set.
	LoginSession          LoginSession `json:",omitzero"`
	RegionData            region.Item
	ScopeProjectId        uuid.ProjectIdNull
	ProjectVpnKeyRegistry map[uuid.ProjectId]entity.VpnKey
}

type LoginSession struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (s LoginSession) IsSet() bool {
	return s.RefreshToken != ""
}
