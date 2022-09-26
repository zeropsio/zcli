//go:build windows

package dns

import (
	"github.com/zerops-io/zcli/src/daemonStorage"
)

func DetectDns() (daemonStorage.LocalDnsManagement, error) {
	return daemonStorage.LocalDnsManagementWindows, nil

}
