package profile

import (
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// dataV1 is the pre-multi-profile global shape. ScopeProjectId lived here directly because the
// CLI only had one layer of persisted state. v2 drops the field (scope is per-repo now); Up()
// either rescues the existing scope into a repo profile or drops it with a hint.
type dataV1 struct {
	versionedjson.Probe
	Region         regionV1           `json:"region"`
	ScopeProjectId uuid.ProjectIdNull `json:"scope_project_id,omitempty"`
	SecretsBackend secrets.Mode       `json:"secrets_backend,omitempty"`
}

func init() {
	globalSchema.Register(1, versionedjson.DecoderFor[dataV1]())
}

// Chain implements versionedjson.Chain[dataChain] so a dataV1 value can be passed to
// versionedjson.Migrate(globalSchema, ...) - the marker family is declared in data.go.
func (dataV1) Chain() dataChain { return dataChain{} }

// Up turns a v1 value into a v2 value. ScopeProjectId is removed from global storage; if a
// scope was set, the migration tries to rescue it into a per-repo profile when cwd is inside
// a git repo without an existing repo profile. When rescue isn't possible the scope is
// dropped with a one-time hint so the user can re-set it explicitly.
func (d dataV1) Up(ctx migrationCtx) (any, error) {
	if d.ScopeProjectId.Filled() {
		rescued, err := ctx.rescueScope(d.ScopeProjectId)
		if err != nil {
			return nil, errors.Wrap(err, "v1 → v2: rescue scoped project into repo profile")
		}
		if !rescued && ctx.log != nil {
			id, _ := d.ScopeProjectId.Get()
			ctx.log("Previously-scoped project " + id.Native() + " was cleared from global state; re-set it inside a repo with `zcli scope project <id>`")
		}
	}
	return dataV2{
		Region:         d.Region,
		SecretsBackend: d.SecretsBackend,
	}, nil
}
