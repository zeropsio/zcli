package profile

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// dataV0 is the on-disk shape before versioning was introduced: default Go field names
// (PascalCase) with no "version" key. Token and ProjectVpnKeyRegistry lived in this file
// directly in plaintext. Legacy stepping stone - never written, only constructed by its
// decoder and immediately upgraded via Up().
type dataV0 struct {
	Token                 string
	RegionData            regionV0
	ScopeProjectId        uuid.ProjectIdNull
	ProjectVpnKeyRegistry map[uuid.ProjectId]vpnKeyV0
}

func init() {
	globalSchema.Register(0, versionedjson.DecoderFor[dataV0]())
}

// Chain implements versionedjson.Chain[dataChain] so a dataV0 value can be passed to
// versionedjson.Migrate(globalSchema, ...) - the marker family is declared in data.go.
func (dataV0) Chain() dataChain { return dataChain{} }

// Up turns a v0 value into a v1 value, draining sensitive fields into the migration context's
// secrets backend along the way. The returned dataV1 is purged of Token and VPN keys; reading
// either from the new shape goes through Manager which delegates to the secrets backend.
func (d dataV0) Up(ctx migrationCtx) (any, error) {
	if d.Token != "" {
		if err := ctx.secrets.Set(secrets.Key(secrets.GlobalScope, "token"), d.Token); err != nil {
			return nil, errors.Wrap(err, "v0 → v1: drain token into secrets backend")
		}
	}
	for projectId, key := range d.ProjectVpnKeyRegistry {
		upgraded := vpnKeyV1{Key: key.Key, ProjectId: key.ProjectId, CreatedAt: key.CreatedAt}
		upgraded.StampVersion(CurrentVpnKeyVersion)
		payload, err := json.Marshal(upgraded)
		if err != nil {
			return nil, errors.Wrapf(err, "v0 → v1: marshal vpn key for project %s", projectId.Native())
		}
		if err := ctx.secrets.Set(secrets.Key(secrets.GlobalScope, "vpn", projectId.Native()), string(payload)); err != nil {
			return nil, errors.Wrapf(err, "v0 → v1: drain vpn key for project %s into secrets backend", projectId.Native())
		}
	}
	return dataV1{
		Region: regionV1{
			Name:       d.RegionData.Name,
			IsDefault:  d.RegionData.IsDefault,
			Address:    d.RegionData.Address,
			GuiAddress: d.RegionData.GuiAddress,
		},
		ScopeProjectId: d.ScopeProjectId,
	}, nil
}
