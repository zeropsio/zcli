// +build darwin

package vpn

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, response *zeropsVpnProtocol.StartVpnResponse) error {
	var err error

	h.logger.Debug("run wireguard-go utun")

	output, err := cmdRunner.Run(exec.Command("wireguard-go", "utun"))
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`INFO: \((.*)\)`)
	submatches := re.FindSubmatch(output)
	if len(submatches) != 2 {
		return errors.New(i18n.VpnStartWireguardInterfaceNotfound)
	}

	interfaceName := string(submatches[1])

	{
		privateKeyName := uuid.New().String()
		tempPrivateKeyFile := path.Join(os.TempDir(), privateKeyName)
		err = ioutil.WriteFile(tempPrivateKeyFile, []byte(privateKey), 0755)
		if err != nil {
			return err
		}
		_, err = cmdRunner.Run(exec.Command("wg", "set", interfaceName, "private-key", tempPrivateKeyFile))
		if err != nil {
			return err
		}
		err = os.Remove(tempPrivateKeyFile)
		if err != nil {
			return err
		}
	}

	_, err = cmdRunner.Run(exec.Command("wg", "set", interfaceName, "listen-port", wireguardPort))
	if err != nil {
		return err
	}

	clientIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := zeropsVpnProtocol.FromProtoIPRange(response.GetVpn().GetVpnIpRange())

	args := []string{
		"set", interfaceName,
		"peer", response.GetVpn().GetServerPublicKey(),
		"allowed-ips", vpnRange.String(),
		"endpoint", selectedVpnAddress + ":" + strconv.Itoa(int(response.GetVpn().GetPort())),
		"persistent-keepalive", "25",
	}
	_, err = cmdRunner.Run(exec.Command("wg", args...))
	if err != nil {
		if !errors.Is(err, cmdRunner.IpAlreadySetErr) {
			panic(err)
		}
	}

	_, err = cmdRunner.Run(exec.Command("ifconfig", interfaceName, "inet6", clientIp.String(), "mtu", "1420"))
	if err != nil {
		return err
	}

	serverIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetServerIp())
	_, err = cmdRunner.Run(exec.Command("route", "add", "-inet6", vpnRange.String(), serverIp.String()))
	if err != nil {
		return err
	}

	return nil
}
