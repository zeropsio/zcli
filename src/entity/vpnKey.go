package entity

import (
	"time"

	"github.com/zeropsio/zerops-go/types/uuid"
)

type VpnKey struct {
	Key       string
	ProjectId uuid.ProjectId
	CreatedAt time.Time
}
