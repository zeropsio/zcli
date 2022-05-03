package dns

import (
	"errors"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/zerops-io/zcli/src/nettools"

	"github.com/zerops-io/zcli/src/dnsServer"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

var UnknownDnsManagementErr = errors.New("unknown dns management")

func SetDns(dnsServer *dnsServer.Handler, dnsIp net.IP, clientIp net.IP, vpnNetwork net.IPNet, dnsManagement LocalDnsManagement) error {
	var err error

	vpnInterfaceName, _, err := nettools.GetInterfaceNameByIp(clientIp)
	if err != nil {
		return err
	}

	switch dnsManagement {
	case LocalDnsManagementUnknown, LocalDnsManagementWindows:
		return nil

	case LocalDnsManagementSystemdResolve:
		_, err = cmdRunner.Run(exec.Command("systemd-resolve", "--set-dns="+dnsIp.String(), `--set-domain=zerops`, "--interface="+vpnInterfaceName))
		if err != nil {
			return err
		}

	case LocalDnsManagementResolveConf:
		err := utils.SetFirstLine(constants.ResolvconfOrderFilePath, "wg*")
		if err != nil {
			return err
		}

		cmd := exec.Command("resolvconf", "-a", vpnInterfaceName)
		cmd.Stdin = strings.NewReader(strings.Join([]string{"nameserver " + dnsIp.String(), "search zerops"}, "\n"))
		_, err = cmdRunner.Run(cmd)
		if err != nil {
			return err
		}

	case LocalDnsManagementFile:
		err := utils.SetFirstLine(constants.ResolvFilePath, "nameserver "+dnsIp.String())
		if err != nil {
			return err
		}

	case LocalDnsManagementScutil:

		var zeropsDynamicStorage ZeropsDynamicStorage
		zeropsDynamicStorage.Read()
		zeropsDynamicStorage.VpnInterfaceName = vpnInterfaceName
		zeropsDynamicStorage.Active = true
		zeropsDynamicStorage.ClientIp = clientIp
		zeropsDynamicStorage.VpnNetwork = vpnNetwork.String()
		zeropsDynamicStorage.DnsIp = dnsIp
		zeropsDynamicStorage.Apply()
		dnsServer.SetAddresses(
			zeropsDynamicStorage.ClientIp,
			zeropsDynamicStorage.ServerAddresses,
			zeropsDynamicStorage.DnsIp,
			vpnNetwork,
		)

	default:
		return UnknownDnsManagementErr
	}

	time.Sleep(3 * time.Second)

	return nil
}
