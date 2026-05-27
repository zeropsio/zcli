package profile

import (
	"encoding/json"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// dataV1 is the current on-disk shape: snake_case json tags and an explicit "version" key.
// When dataV2 lands, add an Up() method here that returns the v2 value, and shift the Data
// type + CurrentDataVersion in handler.go to point at v2.
type dataV1 struct {
	Version               *int                             `json:"version,omitempty"`
	Token                 string                           `json:"token"`
	RegionData            region.Item                      `json:"region_data"`
	ScopeProjectId        uuid.ProjectIdNull               `json:"scope_project_id"`
	ProjectVpnKeyRegistry map[uuid.ProjectId]entity.VpnKey `json:"project_vpn_key_registry"`
}

func init() {
	register(1, func(b []byte) (any, error) {
		var d dataV1
		return d, json.Unmarshal(b, &d)
	})
}
