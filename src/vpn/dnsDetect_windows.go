//go:build windows

package vpn

import (
	"github.com/zeropsio/zcli/src/daemonStorage"
)

func dnsDetect() (daemonStorage.LocalDnsManagement, error) {
	return daemonStorage.LocalDnsManagementWindows, nil

}
