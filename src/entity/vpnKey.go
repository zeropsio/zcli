package entity

import (
	"time"

	"github.com/zeropsio/zerops-go/types/uuid"
)

// VpnKey JSON tags match the un-tagged Go default names (capitalized) so on-disk values
// written by older zcli versions still decode after the linter-driven tag addition. Don't
// rename the tags without a migration.
type VpnKey struct {
	Key       string         `json:"Key"`
	ProjectId uuid.ProjectId `json:"ProjectId"`
	CreatedAt time.Time      `json:"CreatedAt"`
}
