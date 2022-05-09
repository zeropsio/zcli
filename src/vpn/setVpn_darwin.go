//go:build darwin
// +build darwin

package vpn

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	vpnproxy "github.com/zerops-io/zcli/src/proto/vpnproxy"

	"github.com/google/uuid"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

const TunnelNameFile = "/tmp/wg-tun"

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, mtu uint32, response *vpnproxy.StartVpnResponse) error {
	var err error

	h.logger.Debug("run wireguard-go utun")
	cmd := exec.Command("wireguard-go", "utun")
	cmd.Env = []string{"WG_TUN_NAME_FILE=" + TunnelNameFile}
	_, err = cmdRunner.Run(cmd)
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStartWireguardUtunError)
	}

	interfaceName, err := getTunnelName()
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

	clientIp := vpnproxy.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := vpnproxy.FromProtoIPRange(response.GetVpn().GetVpnIpRange())

	args := []string{
		"set", interfaceName,
		"peer", response.GetVpn().GetServerPublicKey(),
		"allowed-ips", vpnRange.String(),
		"endpoint", selectedVpnAddress,
		"persistent-keepalive", "25",
	}
	_, err = cmdRunner.Run(exec.Command("wg", args...))
	if err != nil {
		if !errors.Is(err, cmdRunner.IpAlreadySetErr) {
			panic(err)
		}
	}

	_, err = cmdRunner.Run(exec.Command("ifconfig", interfaceName, "inet6", clientIp.String(), "mtu", strconv.Itoa(int(mtu))))
	if err != nil {
		return err
	}

	serverIp := vpnproxy.FromProtoIP(response.GetVpn().GetServerIp())
	_, err = cmdRunner.Run(exec.Command("route", "add", "-inet6", vpnRange.String(), serverIp.String()))
	if err != nil {
		return err
	}

	return nil
}

func getTunnelName() (string, error) {
	b, err := ioutil.ReadFile(TunnelNameFile)
	if err != nil {
		return "", errors.New(i18n.VpnStartWireguardInterfaceNotfound)
	}
	return strings.TrimSpace(string(b)), nil
}
