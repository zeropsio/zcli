package vpn

import (
	"errors"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/interfaces"

	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

var UnknownDnsManagementErr = errors.New("unknown dns management")

func (h *Handler) setDns(dnsIp net.IP, clientIp net.IP, vpnNetwork net.IPNet, dnsManagement localDnsManagement) error {
	var err error

	vpnInterfaceName, _, err := interfaces.GetInterfaceNameByIp(clientIp)
	if err != nil {
		return err
	}

	switch dnsManagement {
	case localDnsManagementSystemdResolve:
		_, err = cmdRunner.Run(exec.Command("systemd-resolve", "--set-dns="+dnsIp.String(), `--set-domain=zerops`, "--interface="+vpnInterfaceName))
		if err != nil {
			return err
		}

	case localDnsManagementResolveConf:
		err := utils.SetFirstLine(resolvconfOrderFilePath, "wg*")
		if err != nil {
			return err
		}

		cmd := exec.Command("resolvconf", "-a", vpnInterfaceName)
		cmd.Stdin = strings.NewReader("nameserver " + dnsIp.String())
		_, err = cmdRunner.Run(cmd)
		if err != nil {
			return err
		}

	case localDnsManagementFile:
		err := utils.SetFirstLine(resolvFilePath, "nameserver "+dnsIp.String())
		if err != nil {
			return err
		}

	case scutilDnsManagementFile:

		var zeropsDynamicStorage ZeropsDynamicStorage
		zeropsDynamicStorage.Read()
		zeropsDynamicStorage.VpnInterfaceName = vpnInterfaceName
		zeropsDynamicStorage.Active = true
		zeropsDynamicStorage.ClientIp = clientIp
		zeropsDynamicStorage.VpnNetwork = vpnNetwork.String()
		zeropsDynamicStorage.DnsIp = dnsIp
		zeropsDynamicStorage.Apply()
		h.dnsServer.SetAddresses(zeropsDynamicStorage.ClientIp, zeropsDynamicStorage.ServerAddresses, zeropsDynamicStorage.DnsIp, vpnNetwork)

	default:
		return UnknownDnsManagementErr
	}

	time.Sleep(3 * time.Second)

	return nil
}
