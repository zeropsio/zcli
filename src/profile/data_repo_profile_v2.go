package profile

import (
	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// repoProfileV2 is the current per-entry on-disk shape. It adds ScopeClientId on top of
// repoProfileV1: an optional per-repo org override that wins over the global default. The other
// fields keep their repoProfileV1 meaning (Id is the immutable per-profile secrets-scope key,
// Name is what users select on, ScopeProjectId / ScopeServiceId / Region are per-profile state).
type repoProfileV2 struct {
	versionedjson.Probe
	Name           string                  `json:"name"`
	Id             string                  `json:"id"`
	ScopeProjectId uuid.ProjectIdNull      `json:"scope_project_id,omitempty"`
	ScopeServiceId uuid.ServiceStackIdNull `json:"scope_service_id,omitempty"`
	ScopeClientId  uuid.ClientIdNull       `json:"scope_client_id,omitempty"`
	Region         *regionV1               `json:"region,omitempty"`
}

func init() {
	repoProfileSchema.Register(2, versionedjson.DecoderFor[repoProfileV2]())
}

// Chain implements versionedjson.Chain[repoProfileChain] so a repoProfileV2 value can be passed
// to versionedjson.Migrate(repoProfileSchema, ...) - the marker family is declared in
// data_repo_profile.go.
func (repoProfileV2) Chain() repoProfileChain { return repoProfileChain{} }
