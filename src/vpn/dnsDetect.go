package vpn

import (
	"github.com/zeropsio/zcli/src/daemonStorage"
)

func DnsDetect() (daemonStorage.LocalDnsManagement, error) {
	return dnsDetect()
}
