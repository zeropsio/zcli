package profile

import "github.com/zeropsio/zcli/src/versionedjson"

// CurrentVpnKeyVersion is the on-disk version stamp for VPN keys stored in the secrets
// backend. Bumped when the transport struct shape changes.
const CurrentVpnKeyVersion = 1

// VpnKey is the exported view callers consume. Type alias to the latest on-disk shape so
// constructing a VpnKey literal always produces the current target type and bumping the
// version is a single-line flip here. Callers should treat the embedded Probe / Version
// field as an implementation detail - StampVersion handles it on write.
type VpnKey = vpnKeyV1

// vpnKeyCurrent mirrors VpnKey for internal call sites that want the "current target" name
// alongside the schema definition. Kept as an alias of the same underlying type so the two
// stay in lockstep.
type vpnKeyCurrent = vpnKeyV1

// vpnKeySchema dispatches loads on a single VPN key value pulled out of the secrets backend.
// WithMissingVersion(0) covers the legacy path: values written before this chain existed had
// no "version" key and matched Go's default field-name marshalling, which vpnKeyV0 mirrors.
var vpnKeySchema = versionedjson.New[vpnKeyCurrent, migrationCtx](CurrentVpnKeyVersion).WithMissingVersion(0)

// vpnKeyChain is the versionedjson.Chain[C] family marker for every VPN key version. Empty
// struct - only the type identity matters at compile time.
type vpnKeyChain struct{}
