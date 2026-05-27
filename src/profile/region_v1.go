package profile

import (
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/versionedjson"
)

// regionV1 is the current on-disk region shape and - via the Region alias in region.go -
// also the type callers see in the profile API. Carries the "version" key via the embedded
// Probe so the embedded region object self-identifies inside the parent profile JSON.
// Constructors should leave Probe untouched; StampVersion is called on write so the right
// version lands on disk.
type regionV1 struct {
	versionedjson.Probe
	Name       string                `json:"name"`
	IsDefault  bool                  `json:"isDefault"`
	Address    string                `json:"address"`
	GuiAddress optional.Null[string] `json:"guiAddress"`
}

func init() {
	regionSchema.Register(1, versionedjson.DecoderFor[regionV1]())
}

// Chain implements versionedjson.Chain[regionChain] so a regionV1 value can be passed to
// versionedjson.Migrate(regionSchema, ...) - the marker family is declared in region.go.
func (regionV1) Chain() regionChain { return regionChain{} }
