//go:build darwin

package vpn

import (
	"os/exec"

	"github.com/zeropsio/zcli/src/daemonStorage"
)

func dnsDetect() (daemonStorage.LocalDnsManagement, error) {
	if _, err := exec.LookPath("scutil"); err == nil {
		return daemonStorage.LocalDnsManagementNetworkSetup, nil
	}
	return daemonStorage.LocalDnsManagementUnknown, nil

}
