//go:build darwin
// +build darwin

package vpn

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"time"

	vpnproxy "github.com/zerops-io/zcli/src/proto/vpnproxy"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

const TunnelNameFile = "/tmp/wg-zerops"

func (h *Handler) setVpn(ctx context.Context, vpnAddress net.IP, privateKey wgtypes.Key, mtu uint32, response *vpnproxy.StartVpnResponse) error {
	var err error

	h.logger.Debug("run wireguard-go utun")
	cmd := exec.Command("wireguard-go", "utun")
	cmd.Env = []string{"WG_TUN_NAME_FILE=" + TunnelNameFile}
	_, err = cmdRunner.Run(cmd)
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStartWireguardUtunError)
	}

	interfaceName, err := func() (string, error) {
		b, err := ioutil.ReadFile(TunnelNameFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}()
	if err != nil {
		return errors.New(i18n.VpnStartNetworkInterfaceNotFound)
	}

	clientIp := vpnproxy.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := vpnproxy.FromProtoIPRange(response.GetVpn().GetVpnIpRange())
	serverIp := vpnproxy.FromProtoIP(response.GetVpn().GetServerIp())
	serverPublicKey, err := wgtypes.ParseKey(response.GetVpn().ServerPublicKey)
	if err != nil {
		return errors.New(i18n.VpnStartInvalidServerPublicKey)
	}

	udpAddr, ok := netip.AddrFromSlice(vpnAddress)
	if !ok {
		return errors.New(i18n.VpnStartInvalidVpnAddress)
	}
	addr := net.UDPAddrFromAddrPort(
		netip.AddrPortFrom(
			udpAddr,
			uint16(response.GetVpn().GetPort()),
		),
	)

	if err := runCommands(
		ctx,
		h.logger,
		makeCommand(
			"ifconfig",
			i18n.VpnStartUnableToConfigureNetworkInterface,
			interfaceName, "inet6", clientIp.String(), "mtu", strconv.Itoa(int(mtu)),
		),
		makeCommand(
			"route",
			i18n.VpnStartUnableToUpdateRoutingTable,
			"add", "-inet6", vpnRange.String(), serverIp.String(),
		),
	); err != nil {
		return err
	}

	wgClient, err := wgctrl.New()
	if err != nil {
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	defer wgClient.Close()

	keep := time.Second * 25
	if err := wgClient.ConfigureDevice(interfaceName, wgtypes.Config{
		PrivateKey:   &privateKey,
		ListenPort:   nil,
		FirewallMark: nil,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   serverPublicKey,
				Remove:                      false,
				UpdateOnly:                  false,
				PresharedKey:                nil,
				Endpoint:                    addr,
				PersistentKeepaliveInterval: &keep,
				ReplaceAllowedIPs:           false,
				AllowedIPs: []net.IPNet{
					{
						IP:   response.GetVpn().GetVpnIpRange().GetIp(),
						Mask: response.GetVpn().GetVpnIpRange().GetMask(),
					},
				},
			},
		},
	}); err != nil {
		return errors.New(i18n.VpnStartTunnelConfigurationFailed)
	}

	return nil
}
