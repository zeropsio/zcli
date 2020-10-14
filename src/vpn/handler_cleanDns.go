package vpn

import (
	"net"
	"os/exec"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/interfaces"

	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

func (h *Handler) cleanDns(dnsIp, clientIp net.IP, dnsManagement localDnsManagement) error {

	switch dnsManagement {
	case localDnsManagementSystemdResolve:
		return nil
	case localDnsManagementResolveConf:
		vpnInterfaceName, vpnInterfaceFound, err := interfaces.GetInterfaceNameByIp(clientIp)
		if err != nil {
			return err
		}
		if !vpnInterfaceFound {
			return nil
		}
		cmd := exec.Command("resolvconf", "-d", vpnInterfaceName)
		_, err = cmdRunner.Run(cmd)
		if err != nil {
			return err
		}
	case localDnsManagementFile:
		err := utils.RemoveFirstLine(resolvFilePath, "nameserver "+dnsIp.String())
		if err != nil {
			return err
		}
	case scutilDnsManagementFile:
		var zeropsDynamicStorage ZeropsDynamicStorage
		zeropsDynamicStorage.Read()
		zeropsDynamicStorage.Active = false
		zeropsDynamicStorage.Apply()
		h.dnsServer.StopForward()
	default:
		return UnknownDnsManagementErr
	}
	return nil
}
