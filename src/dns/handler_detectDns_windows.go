//go:build windows

package dns

import (
	"github.com/zeropsio/zcli/src/daemonStorage"
)

func DetectDns() (daemonStorage.LocalDnsManagement, error) {
	return daemonStorage.LocalDnsManagementWindows, nil

}
