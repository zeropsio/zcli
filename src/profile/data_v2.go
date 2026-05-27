package profile

import (
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/versionedjson"
)

// dataV2 is the current global on-disk shape. ScopeProjectId moved to RepoProfile (per-repo)
// so this struct holds only state that genuinely belongs to the user globally: their active
// region and their preferred secrets storage backend. Tokens and VPN keys are in the secrets
// backend itself, not in this JSON.
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
