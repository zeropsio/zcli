package profile

import (
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/versionedjson"
)

// repoFileV1 is the pre-client-scope file-level shape: a list of named profile entries plus the
// name of the currently active one. Entries are typed at the static repoProfileV1 element type.
// The wrapper bumped to v2 only because the entry shape did (cascade rule); Up promotes each
// entry through repoProfileSchema so a v1 file lands as a fully v2 file.
//
// Embeds versionedjson.Probe for the wrapper version stamp + StampVersion hook.
type repoFileV1 struct {
	versionedjson.Probe
	ActiveProfile string          `json:"active_profile"`
	Profiles      []repoProfileV1 `json:"profiles"`
}

func init() {
	repoFileSchema.Register(1, versionedjson.DecoderFor[repoFileV1]())
}

// Chain implements versionedjson.Chain[repoFileChain] so a repoFileV1 value can be passed to
// versionedjson.Migrate(repoFileSchema, ...) - the marker family is declared in data_repo.go.
func (repoFileV1) Chain() repoFileChain { return repoFileChain{} }

// Up turns a v1 file into a v2 file. The wrapper shape is unchanged; the work is promoting each
// entry from repoProfileV1 to the current repoProfileV2 via its own schema, so the cascade lands
// at a uniform v2 entry type. Each promoted entry is already version-stamped by Migrate.
func (f repoFileV1) Up(ctx migrationCtx) (any, error) {
	profiles := make([]repoProfileV2, 0, len(f.Profiles))
	for _, entry := range f.Profiles {
		upgraded, _, err := versionedjson.Migrate(repoProfileSchema, entry, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "v1 → v2: migrate repo profile %q", entry.Name)
		}
		profiles = append(profiles, upgraded)
	}
	return repoFileV2{
		ActiveProfile: f.ActiveProfile,
		Profiles:      profiles,
	}, nil
}
