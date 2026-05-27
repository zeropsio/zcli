package profile

import (
	"github.com/zeropsio/zcli/src/versionedjson"
)

// repoFileV1 is the file-level v1 shape: a list of named profile entries plus the name of
// the currently active one. Entries are typed at the static repoProfileV1 element type;
// loadRepo runs each through repoProfileSchema.Migrate so a future older entry shape can be
// promoted forward without bumping the file wrapper.
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
