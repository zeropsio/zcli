//go:build darwin

package dns

import (
	"os/exec"

	"github.com/zeropsio/zcli/src/daemonStorage"
)

func DetectDns() (daemonStorage.LocalDnsManagement, error) {

	if _, err := exec.LookPath("networksetup"); err == nil {
		return daemonStorage.LocalDnsManagementNetworkSetup, nil
	}
	return daemonStorage.LocalDnsManagementUnknown, nil

}
