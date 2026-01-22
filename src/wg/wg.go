package wg

import (
	"net"

	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/dto/output"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type TemplateData struct {
	Mtu                       int
	PrivateKey                string
	PublicKey                 string
	AssignedIpv4Address       string
	AssignedIpv6Address       string
	Ipv4NetworkGateway        string
	ProjectIpv4Network        string
	ProjectIpv6Network        string
	Ipv4Network               string
	Ipv6Network               string
	DnsSetup                  bool
	ProjectIpv4SharedEndpoint string
}

func defaultTemplateData(privateKey wgtypes.Key, vpnSettings output.ProjectVpnItem, mtu int, dnsSetup bool) (TemplateData, error) {
	projectIpv4Network := ""
	if vpnSettings.Project.Ipv4.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Project.Ipv4.Network.Network))
		if err != nil {
			return TemplateData{}, errors.Wrap(err, "failed to parse projectIpv4Network network")
		}
		projectIpv4Network = n.String()
	}

	projectIpv6Network := ""
	if vpnSettings.Project.Ipv6.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Project.Ipv6.Network.Network))
		if err != nil {
			return TemplateData{}, errors.Wrap(err, "failed to parse projectIpv6Network network")
		}
		projectIpv6Network = n.String()
	}

	ipv4Network := ""
	if vpnSettings.Peer.Ipv4.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Peer.Ipv4.Network.Network))
		if err != nil {
			return TemplateData{}, errors.Wrap(err, "failed to parse Ipv4Network network")
		}
		ipv4Network = n.String()
	}

	ipv6Network := ""
	if vpnSettings.Peer.Ipv6.Network.Network != "" {
		_, n, err := net.ParseCIDR(string(vpnSettings.Peer.Ipv6.Network.Network))
		if err != nil {
			return TemplateData{}, errors.Wrap(err, "failed to parse Ipv6Network network")
		}
		ipv6Network = n.String()
	}

	return TemplateData{
		Mtu:                       mtu,
		PrivateKey:                privateKey.String(),
		PublicKey:                 string(vpnSettings.Project.PublicKey),
		AssignedIpv4Address:       string(vpnSettings.Peer.Ipv4.AssignedIpAddress),
		AssignedIpv6Address:       string(vpnSettings.Peer.Ipv6.AssignedIpAddress),
		Ipv4NetworkGateway:        string(vpnSettings.Project.Ipv4.Network.Gateway),
		ProjectIpv4Network:        projectIpv4Network,
		ProjectIpv6Network:        projectIpv6Network,
		Ipv4Network:               ipv4Network,
		Ipv6Network:               ipv6Network,
		DnsSetup:                  dnsSetup,
		ProjectIpv4SharedEndpoint: string(vpnSettings.Project.Ipv4.SharedEndpoint),
	}, nil
}
