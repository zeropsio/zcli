package vpn

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/daemonStorage"
	"github.com/zeropsio/zcli/src/utils"
	"github.com/zeropsio/zcli/src/utils/cmdRunner"
)

var UnknownDnsManagementErr = errors.New("unknown dns management")

func (h *Handler) setDns(ctx context.Context) error {
	data := h.storage.Data()

	switch data.DnsManagement {
	case daemonStorage.LocalDnsManagementUnknown, daemonStorage.LocalDnsManagementWindows:
		return nil

	case daemonStorage.LocalDnsManagementSystemdResolve:
		// resolvectl is multi-binary and behaves differently
		// based on first command argument it receives (name of the command)
		// systemd-resolve is only a symlink to resolvectl
		cmd := exec.Command("resolvectl", "--set-dns="+data.DnsIp.String(), `--set-domain=zerops`, "--interface="+data.InterfaceName)
		cmd.Args[0] = "systemd-resolve"
		if _, err := cmdRunner.Run(cmd); err != nil {
			return err
		}

	case daemonStorage.LocalDnsManagementResolveConf:
		err := utils.SetFirstLine(constants.ResolvconfOrderFilePath, "wg*")
		if err != nil {
			return err
		}

		cmd := exec.Command("resolvconf", "-a", data.InterfaceName)
		cmd.Stdin = strings.NewReader(strings.Join([]string{"nameserver " + data.DnsIp.String(), "search zerops"}, "\n"))
		if _, err = cmdRunner.Run(cmd); err != nil {
			return err
		}

	case daemonStorage.LocalDnsManagementFile:
		if err := utils.SetFirstLine(constants.ResolvFilePath, "nameserver "+data.DnsIp.String()); err != nil {
			return err
		}

	case daemonStorage.LocalDnsManagementNetworkSetup:

		if err := h.setDnsNetworksetup(ctx); err != nil {
			return err
		}

	default:
		return UnknownDnsManagementErr
	}

	return nil
}
