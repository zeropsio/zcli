package profile

import (
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/versionedjson"
)

// dataV2 is the pre-client-scope global shape. ScopeProjectId moved to RepoProfile (per-repo)
// so this struct holds only the user's active region and preferred secrets storage backend.
// Tokens and VPN keys are in the secrets backend itself, not in this JSON. v3 adds an optional
// global client scope.
//
// Embeds versionedjson.Probe to inherit both the "version" JSON field and the StampVersion
// auto-stamp hook that Schema.Load calls after migration.
type dataV2 struct {
	versionedjson.Probe
	Region         regionV1     `json:"region"`
	SecretsBackend secrets.Mode `json:"secrets_backend,omitempty"`
}

func init() {
	globalSchema.Register(2, versionedjson.DecoderFor[dataV2]())
}

// Chain implements versionedjson.Chain[dataChain] so a dataV2 value can be passed to
// versionedjson.Migrate(globalSchema, ...) - the marker family is declared in data.go.
func (dataV2) Chain() dataChain { return dataChain{} }

// Up turns a v2 value into a v3 value. The only delta is the new ScopeClientId field, which
// defaults to empty (no global org scope) - existing fields carry over unchanged.
func (d dataV2) Up(_ migrationCtx) (any, error) {
	return dataV3{
		Region:         d.Region,
		SecretsBackend: d.SecretsBackend,
	}, nil
}
