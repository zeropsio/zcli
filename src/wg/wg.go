package wg

import (
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/dto/output"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func defaultTemplateData(privateKey wgtypes.Key, vpnSettings output.ProjectVpnItem, mtu int) (map[string]string, error) {
	projectIpv4Network := ""
	if vpnSettings.Project.Ipv4.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Project.Ipv4.Network.Network))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse projectIpv4Network network")
		}
		projectIpv4Network = n.String()
	}

	projectIpv6Network := ""
	if vpnSettings.Project.Ipv6.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Project.Ipv6.Network.Network))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse projectIpv6Network network")
		}
		projectIpv6Network = n.String()
	}

	ipv4Network := ""
	if vpnSettings.Peer.Ipv4.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Peer.Ipv4.Network.Network))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Ipv4Network network")
		}
		ipv4Network = n.String()
	}

	ipv6Network := ""
	if vpnSettings.Peer.Ipv6.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Peer.Ipv6.Network.Network))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Ipv6Network network")
		}
		ipv6Network = n.String()
	}

	return map[string]string{
		"Mtu":                       strconv.Itoa(mtu),
		"PrivateKey":                privateKey.String(),
		"PublicKey":                 string(vpnSettings.Project.PublicKey),
		"AssignedIpv4Address":       string(vpnSettings.Peer.Ipv4.AssignedIpAddress),
		"AssignedIpv6Address":       string(vpnSettings.Peer.Ipv6.AssignedIpAddress),
		"Ipv4NetworkGateway":        string(vpnSettings.Project.Ipv4.Network.Gateway),
		"ProjectIpv4Network":        projectIpv4Network,
		"ProjectIpv6Network":        projectIpv6Network,
		"Ipv4Network":               ipv4Network,
		"Ipv6Network":               ipv6Network,
		"ProjectIpv4SharedEndpoint": string(vpnSettings.Project.Ipv4.SharedEndpoint),
	}, nil
}
