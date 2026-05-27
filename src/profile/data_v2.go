package profile

import (
	"encoding/json"

	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/secrets"
)

// dataV2 is the current global on-disk shape. ScopeProjectId moved to RepoData (per-repo) so
// this struct holds only state that genuinely belongs to the user globally: their active
// region and their preferred secrets storage backend. Tokens and VPN keys are in the secrets
// backend itself, not in this JSON.
type dataV2 struct {
	Version        *int         `json:"version,omitempty"`
	Region         region.Item  `json:"region"`
	SecretsBackend secrets.Mode `json:"secrets_backend,omitempty"`
}

func init() {
	register(2, func(b []byte) (any, error) {
		var d dataV2
		return d, json.Unmarshal(b, &d)
	})
}
