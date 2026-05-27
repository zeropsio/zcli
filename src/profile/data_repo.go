package profile

import "github.com/zeropsio/zcli/src/versionedjson"

// CurrentRepoVersion is the on-disk version stamp for the repo profile file wrapper. Lives
// next to repoFileV1 (the current target) by convention - the latest version file owns the
// constant. Bumped when the file-level shape itself changes OR when the child entry shape
// bumps (because each repoFileVN statically holds []repoProfileVN; bumping the child without
// bumping this wrapper would leave the wrapper's static element type pinned to the older
// child shape - see "Cascading" in handler.go's package doc).
const CurrentRepoVersion = 1

// DefaultRepoProfileName is the name assigned to the sole profile produced by `zcli profile
// init`, and the suggested name for the first entry users keep around.
const DefaultRepoProfileName = "default"

// RepoFile is the in-memory + write-side shape of the repo profile file. Type alias to
// repoFileV1 (the current target) so bumping the version is a single-line flip; the
// embedded versionedjson.Probe carries the auto-stamp hook through.
type RepoFile = repoFileV1

// repoFileSchema dispatches loads on the file wrapper version (active_profile + profiles[]).
// Entries decode straight into typed repoProfileV1 values; each is then run through
// repoProfileSchema for its own version dispatch so the two chains evolve independently.
var repoFileSchema = versionedjson.New[RepoFile, migrationCtx](CurrentRepoVersion)

// repoFileChain is the versionedjson.Chain[C] family marker for every repo-file-wrapper
// version. Every repoFileVN implements Chain() returning this so
// versionedjson.Migrate(repoFileSchema, ...) rejects cross-family values at compile time.
type repoFileChain struct{}
