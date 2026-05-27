package profile

import (
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// CurrentRepoVersion is the on-disk version stamp for repoData. Bumped only when the repo
// profile schema actually changes; on first revision there is no migration to write.
const CurrentRepoVersion = 1

// repoDataV1 is the current per-repo profile shape, written to <repo>/.zcli/profile.json.
// Holds non-sensitive per-repo state only. Optional fields (Region) override the global
// equivalent when present; ScopeProjectId is repo-only and has no global counterpart.
//
// Id is a UUID generated when the repo profile is initialised and never changes thereafter -
// it scopes per-repo secrets in the secrets backend (e.g. repo/<id>/token) so a token entered
// for this repo on this machine doesn't collide with the global token or another repo's.
type repoDataV1 struct {
	Version        *int               `json:"version,omitempty"`
	Id             string             `json:"id"`
	ScopeProjectId uuid.ProjectIdNull `json:"scope_project_id,omitempty"`
	Region         *region.Item       `json:"region,omitempty"`
}

// repoCurrent aliases the latest repoDataVN. Single-version chain for now - when a repoDataV2
// lands, flip this alias the same way as dataCurrent in handler.go.
type repoCurrent = repoDataV1

// RepoData is the distinct type exposed to callers, mirroring the Data/dataCurrent split for
// global profile data. Methods on RepoData (Marshal/Unmarshal once needed) hang here, not on
// the underlying versioned struct.
type RepoData repoCurrent
