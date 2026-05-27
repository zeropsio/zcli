package profile

import (
	"time"

	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// vpnKeyV1 is the current on-disk transport shape and - via the VpnKey alias in vpn_key.go -
// also the type callers see in the public Manager API. Carries the "version" key via the
// embedded Probe and uses camelCase JSON tags. Constructors should leave Probe untouched;
// StampVersion is called on write so the right version lands on disk.
type vpnKeyV1 struct {
	versionedjson.Probe
	Key       string         `json:"key"`
	ProjectId uuid.ProjectId `json:"projectId"`
	CreatedAt time.Time      `json:"createdAt"`
}

func init() {
	vpnKeySchema.Register(1, versionedjson.DecoderFor[vpnKeyV1]())
}

// Chain implements versionedjson.Chain[vpnKeyChain] so a vpnKeyV1 value can be passed to
// versionedjson.Migrate(vpnKeySchema, ...) - the marker family is declared in vpn_key.go.
func (vpnKeyV1) Chain() vpnKeyChain { return vpnKeyChain{} }
