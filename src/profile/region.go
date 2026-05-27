package profile

import (
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/versionedjson"
)

// CurrentRegionVersion is the on-disk version stamp for the region value persisted inside
// the global profile and (optionally) each repo profile entry. Bumped when the embedded
// region shape changes; today both versions hold the same data, the chain is wired so a
// future field change can migrate cleanly.
const CurrentRegionVersion = 1

// Region is the exported view callers consume and the latest on-disk shape. Aliased to
// regionV1 so bumping the version is a single-line flip here. region.Item (the API DTO
// returned by the regions endpoint) is intentionally a separate type - callers convert with
// RegionFromItem at the API boundary so the wire format and the on-disk format can evolve
// independently.
type Region = regionV1

// regionSchema dispatches loads on a standalone region value. The profile data paths usually
// decode region through json.Unmarshal as part of the parent (dataVN, repoProfileVN), but the
// schema is the canonical place to register decoders + Up steps so the chain stays in shape
// for future field changes.
var regionSchema = versionedjson.New[Region, migrationCtx](CurrentRegionVersion).WithMissingVersion(0)

// regionChain is the versionedjson.Chain[C] family marker for every region version. Empty
// struct - only the type identity matters at compile time.
type regionChain struct{}

// RegionFromItem converts the API DTO returned by region.Handler.RetrieveAllFromURL into
// the on-disk shape. Stamps the current version so writers don't have to remember.
func RegionFromItem(it region.Item) Region {
	r := Region{
		Name:       it.Name,
		IsDefault:  it.IsDefault,
		Address:    it.Address,
		GuiAddress: it.GuiAddress,
	}
	r.StampVersion(CurrentRegionVersion)
	return r
}
