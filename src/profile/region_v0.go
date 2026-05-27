package profile

import (
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/versionedjson"
)

// regionV0 is the legacy on-disk region shape: identical field names and JSON tags to the
// current target, but no embedded Probe (so values written before this chain existed have
// no "version" key inside the region object). Pre-versioning data on disk decodes into v0
// via the schema's missing-version default; Up() re-tags into the current shape.
type regionV0 struct {
	Name       string                `json:"name"`
	IsDefault  bool                  `json:"isDefault"`
	Address    string                `json:"address"`
	GuiAddress optional.Null[string] `json:"guiAddress"`
}

func init() {
	regionSchema.Register(0, versionedjson.DecoderFor[regionV0]())
}

// Chain implements versionedjson.Chain[regionChain] so a regionV0 value can be passed to
// versionedjson.Migrate(regionSchema, ...) - the marker family is declared in region.go.
func (regionV0) Chain() regionChain { return regionChain{} }

// Up turns a v0 value into the current shape. Pure re-tag - no data transformation.
func (r regionV0) Up(_ migrationCtx) (any, error) {
	return regionV1{
		Name:       r.Name,
		IsDefault:  r.IsDefault,
		Address:    r.Address,
		GuiAddress: r.GuiAddress,
	}, nil
}
