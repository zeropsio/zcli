// +build darwin

package startVpn

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"

	"github.com/zerops-io/zcli/src/service/sudoers"

	"github.com/google/uuid"
	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, response *zeropsVpnProtocol.StartVpnResponse) error {
	var err error

	output, err := h.sudoers.RunCommand(exec.Command("wireguard-go", "utun"))
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`INFO: \((.*)\)`)
	submatches := re.FindSubmatch(output)
	if len(submatches) != 2 {
		return errors.New("vpn interface not found")
	}

	interfaceName := string(submatches[1])

	{
		privateKeyName := uuid.New().String()
		tempPrivateKeyFile := path.Join(os.TempDir(), privateKeyName)

		fmt.Println(tempPrivateKeyFile)
		err = ioutil.WriteFile(tempPrivateKeyFile, []byte(privateKey), 0755)
		if err != nil {
			return err
		}
		_, err = h.sudoers.RunCommand(exec.Command("wg", "set", interfaceName, "private-key", tempPrivateKeyFile))
		if err != nil {
			return err
		}
		err = os.Remove(tempPrivateKeyFile)
		if err != nil {
			return err
		}
	}

	_, err = h.sudoers.RunCommand(exec.Command("wg", "set", interfaceName, "listen-port", wireguardPort))
	if err != nil {
		return err
	}

	clientIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	serverIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetServerIp())
	vpnRange := zeropsVpnProtocol.FromProtoIPRange(response.GetVpn().GetVpnIpRange())

	args := []string{
		"set", interfaceName,
		"peer", response.GetVpn().GetServerPublicKey(),
		"allowed-ips", vpnRange.String(),
		"endpoint", selectedVpnAddress + ":" + strconv.Itoa(int(response.GetVpn().GetPort())),
		"persistent-keepalive", "25",
	}
	_, err = h.sudoers.RunCommand(exec.Command("wg", args...))
	if err != nil {
		if !errors.Is(err, sudoers.IpAlreadySetErr) {
			panic(err)
		}
	}

	_, err = h.sudoers.RunCommand(exec.Command("ifconfig", interfaceName, "inet6", clientIp.String(), "mtu", "1420"))
	if err != nil {
		return err
	}

	_, err = h.sudoers.RunCommand(exec.Command("route", "add", "-inet6", vpnRange.String(), serverIp.String()))
	if err != nil {
		return err
	}

	h.logger.Debug("assigned client address: " + clientIp.String())
	h.logger.Debug("assigned vpn server: " + selectedVpnAddress + ":" + strconv.Itoa(int(response.GetVpn().GetPort())))
	h.logger.Debug("server public key: " + response.GetVpn().GetServerPublicKey())
	h.logger.Debug("serverIp address: " + serverIp.String())
	h.logger.Debug("vpnRange: " + vpnRange.String())

	h.storage.Data.ServerIp = serverIp.String()

	return nil
}
