package profile

import (
	"encoding/json"

	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// dataV1 is the current on-disk shape. Holds non-sensitive user state only - tokens and VPN
// keys live in the secrets backend (see Manager.Token, Manager.VpnKey) and never appear in this
// JSON. SecretsBackend records the user's preferred storage mode so subsequent CLI invocations
// stay consistent without needing the env var to be set every time.
type dataV1 struct {
	Version        *int               `json:"version,omitempty"`
	Region         region.Item        `json:"region"`
	ScopeProjectId uuid.ProjectIdNull `json:"scope_project_id,omitempty"`
	SecretsBackend secrets.Mode       `json:"secrets_backend,omitempty"`
}

func init() {
	register(1, func(b []byte) (any, error) {
		var d dataV1
		return d, json.Unmarshal(b, &d)
	})
}
