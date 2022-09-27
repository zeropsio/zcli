//go:build linux
// +build linux

package vpn

import (
	"context"
	_ "embed"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/lxc/lxd/shared/logger"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/vpnproxy"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const interfaceName = "zerops"

func (h *Handler) setVpn(ctx context.Context, vpnAddress net.IP, privateKey wgtypes.Key, mtu uint32, response *vpnproxy.StartVpnResponse) error {

	clientIp := vpnproxy.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := vpnproxy.FromProtoIPRange(response.GetVpn().GetVpnIpRange())
	serverPublicKey, err := wgtypes.ParseKey(response.GetVpn().ServerPublicKey)
	if err != nil {
		return err
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

	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, in := range interfaces {
		if in.Name == interfaceName {
			if err := runCommands(
				ctx,
				h.logger,
				makeCommand(
					"ip",
					i18n.VpnStopUnableToRemoveTunnelInterface,
					"link", "del", interfaceName,
				),
			); err != nil {
				return err
			}
		}
	}

	if err := runCommands(
		ctx,
		h.logger,
		makeCommand(
			"ip",
			i18n.VpnStartUnableToConfigureNetworkInterface,
			"link", "add", interfaceName, "type", "wireguard",
		),
		makeCommand(
			"ip",
			i18n.VpnStartUnableToConfigureNetworkInterface,
			"-6", "address", "add", clientIp.String()+"/128", "dev", interfaceName,
		),
		makeCommand(
			"ip",
			i18n.VpnStartUnableToConfigureNetworkInterface,
			"link", "set", "dev", interfaceName, "mtu", strconv.Itoa(int(mtu)), "up",
		),
		makeCommand(
			"ip",
			i18n.VpnStartUnableToUpdateRoutingTable,
			"route", "add", vpnRange.String(), "dev", interfaceName,
		),
	); err != nil {
		return err
	}

	wgClient, err := wgctrl.New()
	if err != nil {
		logger.Error(err.Error())
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
		logger.Error(err.Error())
		return errors.New(i18n.VpnStartTunnelConfigurationFailed)
	}

	return nil
}
