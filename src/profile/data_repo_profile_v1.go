package profile

import (
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/versionedjson"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// repoProfileV1 is the current per-entry on-disk shape. Each entry is decoded through
// repoProfileSchema during file load, so an older entry shape would Up itself into this
// struct just like any other versioned document.
//
// Id is a UUID generated when the entry is created and the per-profile secrets-scope key -
// never reassigned. Name is what users select on. ScopeProjectId / Region are the per-profile
// state pieces the rest of the Manager reads. Embeds versionedjson.Probe to inherit the
// "version" JSON field and StampVersion hook.
type repoProfileV1 struct {
	versionedjson.Probe
	Name           string             `json:"name"`
	Id             string             `json:"id"`
	ScopeProjectId uuid.ProjectIdNull `json:"scope_project_id,omitempty"`
	Region         *region.Item       `json:"region,omitempty"`
}

func init() {
	repoProfileSchema.Register(1, versionedjson.DecoderFor[repoProfileV1]())
}

// Chain implements versionedjson.Chain[repoProfileChain] so a repoProfileV1 value can be
// passed to versionedjson.Migrate(repoProfileSchema, ...) - the marker family is declared
// alongside the schema in data_repo_profile.go.
func (repoProfileV1) Chain() repoProfileChain { return repoProfileChain{} }
