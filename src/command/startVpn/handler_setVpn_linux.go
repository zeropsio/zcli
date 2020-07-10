// +build linux

package startVpn

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/google/uuid"
	"github.com/zerops-io/zcli/src/service/sudoers"
	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, response *zeropsVpnProtocol.StartVpnResponse) error {
	var err error

	_, err = h.sudoers.RunCommand(exec.Command("ip", "link", "add", "wg0", "type", "wireguard"))
	if err != nil {
		if !errors.Is(err, sudoers.IpAlreadySetErr) {
			return err
		}
	}

	_, err = h.sudoers.RunCommand(exec.Command("ip", "link", "set", "mtu", "1420", "up", "dev", "wg0"))
	if err != nil {
		return err
	}

	{
		privateKeyName := uuid.New().String()
		tempPrivateKeyFile := path.Join(os.TempDir(), privateKeyName)
		err = ioutil.WriteFile(tempPrivateKeyFile, []byte(privateKey), 0755)
		if err != nil {
			return err
		}
		_, err = h.sudoers.RunCommand(exec.Command("wg", "set", "wg0", "private-key", tempPrivateKeyFile))
		if err != nil {
			return err
		}
		err = os.Remove(tempPrivateKeyFile)
		if err != nil {
			return err
		}
	}

	_, err = h.sudoers.RunCommand(exec.Command("ip", "link", "set", "wg0", "up"))
	if err != nil {
		return err
	}

	_, err = h.sudoers.RunCommand(exec.Command("wg", "set", "wg0", "listen-port", wireguardPort))
	if err != nil {
		return err
	}

	clientIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	serverIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetServerIp())
	vpnRange := zeropsVpnProtocol.FromProtoIPRange(response.GetVpn().GetVpnIpRange())

	args := []string{
		"set", "wg0",
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

	_, err = h.sudoers.RunCommand(exec.Command("ip", "-6", "address", "add", clientIp.String(), "dev", "wg0"))
	if err != nil {
		return err
	}

	_, err = h.sudoers.RunCommand(exec.Command("ip", "route", "add", vpnRange.String(), "dev", "wg0"))
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
