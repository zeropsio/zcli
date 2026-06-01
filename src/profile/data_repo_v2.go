package profile

import (
	"github.com/zeropsio/zcli/src/versionedjson"
)

// repoFileV2 is the file-level v2 shape: same wrapper as v1 (active profile + entry list) but
// its entries are statically typed at repoProfileV2. The wrapper version only bumped because
// the child entry shape did - see the "Cascading" rule in data_repo.go / handler.go's package
// doc. loadRepo runs each entry through repoProfileSchema.Migrate inside repoFileV1.Up, so an
// older entry shape is promoted forward before it lands here.
type repoFileV2 struct {
	versionedjson.Probe
	ActiveProfile string          `json:"active_profile"`
	Profiles      []repoProfileV2 `json:"profiles"`
}

func init() {
	repoFileSchema.Register(2, versionedjson.DecoderFor[repoFileV2]())
}

// Chain implements versionedjson.Chain[repoFileChain] so a repoFileV2 value can be passed to
// versionedjson.Migrate(repoFileSchema, ...) - the marker family is declared in data_repo.go.
func (repoFileV2) Chain() repoFileChain { return repoFileChain{} }
