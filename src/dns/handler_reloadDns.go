package dns

import (
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dnsServer"
)

func ReloadDns(data daemonStorage.Data, dns *dnsServer.Handler) error {
	switch data.DnsManagement {
	case
		daemonStorage.LocalDnsManagementUnknown,
		daemonStorage.LocalDnsManagementWindows,
		daemonStorage.LocalDnsManagementSystemdResolve,
		daemonStorage.LocalDnsManagementResolveConf,
		daemonStorage.LocalDnsManagementFile:

		return nil

	case
		daemonStorage.LocalDnsManagementNetworkSetup,
		daemonStorage.LocalDnsManagementScutil:

		if err := setDnsByNetworksetup(data, dns, data.InterfaceName != ""); err != nil {
			return err
		}
	default:
		return UnknownDnsManagementErr
	}

	return nil
}
