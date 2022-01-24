package vpn

import (
	"strconv"

	"github.com/zerops-io/zcli/src/wgquick"

	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, mtu uint32, response *zeropsVpnProtocol.StartVpnResponse) error {
	clientIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := zeropsVpnProtocol.FromProtoIPRange(response.GetVpn().GetVpnIpRange())
	dnsIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetDnsIp())
	serverAddress := selectedVpnAddress + ":" + strconv.Itoa(int(response.GetVpn().GetPort()))

	err := wgquick.New().Up("zerops", wgquick.Config{
		ClientAddress:    clientIp,
		ServerPublicKey:  response.GetVpn().GetServerPublicKey(),
		DnsServers:       []string{dnsIp.String(), "zerops"},
		ServerAddress:    serverAddress,
		AllowedIPs:       vpnRange,
		ClientPrivateKey: privateKey,
	})

	return err
}
