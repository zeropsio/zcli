package profile

import (
	"encoding/json"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// dataV0 is the on-disk shape before versioning was introduced: default Go field names
// (PascalCase) with no "version" key. Legacy stepping stone for migrations - never written,
// only constructed by its decoder and immediately upgraded via Up().
type dataV0 struct {
	Token                 string
	RegionData            region.Item
	ScopeProjectId        uuid.ProjectIdNull
	ProjectVpnKeyRegistry map[uuid.ProjectId]entity.VpnKey
}

func init() {
	register(0, func(b []byte) (any, error) {
		var d dataV0
		return d, json.Unmarshal(b, &d)
	})
}

// Up turns a v0 value into a v1 value. v0 → v1 is a field-naming change only; copy fields.
func (d dataV0) Up() any {
	return dataV1{
		Token:                 d.Token,
		RegionData:            d.RegionData,
		ScopeProjectId:        d.ScopeProjectId,
		ProjectVpnKeyRegistry: d.ProjectVpnKeyRegistry,
	}
}
