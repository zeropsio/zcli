package dns

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dnsServer"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

var UnknownDnsManagementErr = errors.New("unknown dns management")

func SetDns(data daemonStorage.Data, dns *dnsServer.Handler) error {
	var err error

	switch data.DnsManagement {
	case daemonStorage.LocalDnsManagementUnknown, daemonStorage.LocalDnsManagementWindows:
		return nil

	case daemonStorage.LocalDnsManagementSystemdResolve:
		// resolvectl is multi-binary and behaves differently
		// based on first command argument it receives (name of the command)
		// systemd-resolve is only a symlink to resolvectl
		cmd := exec.Command("resolvectl", "--set-dns="+data.DnsIp.String(), `--set-domain=zerops`, "--interface="+data.InterfaceName)
		cmd.Args[0] = "systemd-resolve"
		_, err = cmdRunner.Run(cmd)
		if err != nil {
			return err
		}

	case daemonStorage.LocalDnsManagementResolveConf:
		err := utils.SetFirstLine(constants.ResolvconfOrderFilePath, "wg*")
		if err != nil {
			return err
		}

		cmd := exec.Command("resolvconf", "-a", data.InterfaceName)
		cmd.Stdin = strings.NewReader(strings.Join([]string{"nameserver " + data.DnsIp.String(), "search zerops"}, "\n"))
		_, err = cmdRunner.Run(cmd)
		if err != nil {
			return err
		}

	case daemonStorage.LocalDnsManagementFile:
		err := utils.SetFirstLine(constants.ResolvFilePath, "nameserver "+data.DnsIp.String())
		if err != nil {
			return err
		}

	case
		daemonStorage.LocalDnsManagementNetworkSetup,
		daemonStorage.LocalDnsManagementScutil:

		if err := setDnsByNetworksetup(data, dns, true); err != nil {
			return err
		}

	default:
		return UnknownDnsManagementErr
	}

	time.Sleep(3 * time.Second)

	return nil
}
