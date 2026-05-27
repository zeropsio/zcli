package profile

import (
	"time"

	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// vpnKeyV0 is the un-versioned legacy on-disk shape: Go default field names, no "version"
// key. Produced by older zcli versions that marshalled the VPN key struct without JSON tags;
// also embedded directly in the legacy plaintext profile (dataV0.ProjectVpnKeyRegistry).
// Only constructed by its decoder or by reading the legacy profile, and immediately upgraded
// via Up().
type vpnKeyV0 struct {
	Key       string
	ProjectId uuid.ProjectId
	CreatedAt time.Time
}

func init() {
	vpnKeySchema.Register(0, versionedjson.DecoderFor[vpnKeyV0]())
}

// Chain implements versionedjson.Chain[vpnKeyChain] so a vpnKeyV0 value can be passed to
// versionedjson.Migrate(vpnKeySchema, ...) - the marker family is declared in vpn_key.go.
func (vpnKeyV0) Chain() vpnKeyChain { return vpnKeyChain{} }

// Up turns a v0 value into the current shape. Pure re-tag - no data transformation.
func (k vpnKeyV0) Up(_ migrationCtx) (any, error) {
	return vpnKeyV1{
		Key:       k.Key,
		ProjectId: k.ProjectId,
		CreatedAt: k.CreatedAt,
	}, nil
}
