package profile

import "github.com/zeropsio/zcli/src/versionedjson"

// CurrentRepoProfileVersion is the on-disk version stamp for a single profile entry inside
// the repo file. The two chains migrate independently at the schema level, but bumping this
// constant requires also bumping the file wrapper to a new repoFileV(N+1) whose Profiles
// slot is []repoProfileV(N+1) - otherwise the parent's static element type still pins to
// the older child shape and new fields can't round-trip. See "Cascading" in handler.go's
// package doc for the full rationale.
const CurrentRepoProfileVersion = 2

// RepoProfile is the exported view of a single profile entry. Type alias rather than a
// distinct type so callers like `profile.RepoProfile{Id: ..., ScopeProjectId: ...}` work
// without conversions and always get the latest entry shape via the repoProfileV* alias.
type RepoProfile = repoProfileV2

// repoProfileSchema dispatches loads on a single entry inside the repo file. Each entry in
// the profiles[] array is fed through this schema individually during loadRepo so an older
// entry shape can be migrated forward without a file-level version bump.
var repoProfileSchema = versionedjson.New[RepoProfile, migrationCtx](CurrentRepoProfileVersion)

// repoProfileChain is the versionedjson.Chain[C] family marker for every per-entry version.
// Today that's just repoProfileV1; future repoProfileV2/V3/... return the same marker so
// versionedjson.Migrate type-checks them against repoProfileSchema. Empty struct - only the
// type identity matters.
type repoProfileChain struct{}
