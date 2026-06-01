package profile

import (
	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// repoProfileV1 is the pre-client-scope per-entry on-disk shape. Each entry is decoded through
// repoProfileSchema during file load, so an older entry shape would Up itself into the current
// struct just like any other versioned document. v2 adds an optional per-repo ScopeClientId.
//
// Id is a UUID generated when the entry is created and the per-profile secrets-scope key -
// never reassigned. Name is what users select on. ScopeProjectId / Region are the per-profile
// state pieces the rest of the Manager reads. Embeds versionedjson.Probe to inherit the
// "version" JSON field and StampVersion hook.
type repoProfileV1 struct {
	versionedjson.Probe
	Name           string                  `json:"name"`
	Id             string                  `json:"id"`
	ScopeProjectId uuid.ProjectIdNull      `json:"scope_project_id,omitempty"`
	ScopeServiceId uuid.ServiceStackIdNull `json:"scope_service_id,omitempty"`
	Region         *regionV1               `json:"region,omitempty"`
}

func init() {
	repoProfileSchema.Register(1, versionedjson.DecoderFor[repoProfileV1]())
}

// Chain implements versionedjson.Chain[repoProfileChain] so a repoProfileV1 value can be
// passed to versionedjson.Migrate(repoProfileSchema, ...) - the marker family is declared
// alongside the schema in data_repo_profile.go.
func (repoProfileV1) Chain() repoProfileChain { return repoProfileChain{} }

// Up turns a v1 entry into a v2 entry. The only delta is the new ScopeClientId field, which
// defaults to empty (no per-repo org override) - existing fields carry over unchanged.
func (p repoProfileV1) Up(_ migrationCtx) (any, error) {
	return repoProfileV2{
		Name:           p.Name,
		Id:             p.Id,
		ScopeProjectId: p.ScopeProjectId,
		ScopeServiceId: p.ScopeServiceId,
		Region:         p.Region,
	}, nil
}
