package profile

import "github.com/zeropsio/zcli/src/versionedjson"

// CurrentDataVersion is the on-disk version stamp for the global profile JSON. Lives next to
// dataV2 (the current target) by convention - the latest version file owns the constant.
const CurrentDataVersion = 2

// Data is the current global on-disk shape. Type alias - methods (e.g. StampVersion) hang on
// the underlying versioned struct so save/load paths don't need explicit conversions.
type Data = dataV2

// globalSchema is the versioned-load schema for the global profile JSON file. Decoders for
// each dataVN register on this in their init() blocks; Manager.loadGlobal invokes Load to walk
// older shapes up to dataCurrent. Missing-version is 0 because the pre-versioning v0 shape
// didn't stamp a "version" key.
var globalSchema = versionedjson.New[Data, migrationCtx](CurrentDataVersion).
	WithMissingVersion(0)

// dataChain is the versionedjson.Chain[C] family marker for every global-profile version.
// Every dataVN implements Chain() returning this so versionedjson.Migrate(globalSchema, ...)
// rejects values from a different family at compile time.
type dataChain struct{}
