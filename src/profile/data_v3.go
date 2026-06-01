package profile

import (
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// dataV3 is the current global on-disk shape. It adds ScopeClientId on top of dataV2: an
// optional global default org. Unlike the per-repo project/service scope, client scope is
// allowed at the global layer and merges field-level with the repo layer (repo wins), the same
// way Region does - see Manager.ScopeClient.
//
// Embeds versionedjson.Probe to inherit both the "version" JSON field and the StampVersion
// auto-stamp hook that Schema.Load calls after migration.
type dataV3 struct {
	versionedjson.Probe
	Region         regionV1          `json:"region"`
	SecretsBackend secrets.Mode      `json:"secrets_backend,omitempty"`
	ScopeClientId  uuid.ClientIdNull `json:"scope_client_id,omitempty"`
}

func init() {
	globalSchema.Register(3, versionedjson.DecoderFor[dataV3]())
}

// Chain implements versionedjson.Chain[dataChain] so a dataV3 value can be passed to
// versionedjson.Migrate(globalSchema, ...) - the marker family is declared in data.go.
func (dataV3) Chain() dataChain { return dataChain{} }
