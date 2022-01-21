//go:build windows
// +build windows

package vpn

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

const Template = `
[Interface]
PrivateKey = {{.ClientPrivateKey}}
Address = {{.ClientAddress}}
DNS = {{.DnsServers}}

[Peer]
PublicKey = {{.ServerPublicKey}}
AllowedIPs = {{.AllowedIPs}}
Endpoint = {{.ServerAddress}}
`

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, mtu uint32, response *zeropsVpnProtocol.StartVpnResponse) error {
	_, err := exec.LookPath("wireguard")
	if err != nil {
		return err
	}

	clientIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := zeropsVpnProtocol.FromProtoIPRange(response.GetVpn().GetVpnIpRange())
	dnsIp := zeropsVpnProtocol.FromProtoIP(response.GetVpn().GetDnsIp())
	serverAddress := selectedVpnAddress + ":" + strconv.Itoa(int(response.GetVpn().GetPort()))

	tmpl := template.Must(template.New("").Parse(strings.Replace(Template, "\n", "\r\n", -1)))

	var bf bytes.Buffer
	err = tmpl.Execute(&bf, struct {
		ClientPrivateKey string
		ClientAddress    string
		DnsServers       string

		ServerPublicKey string
		AllowedIPs      string
		ServerAddress   string
	}{
		ClientPrivateKey: privateKey,
		ClientAddress:    clientIp.String(),
		AllowedIPs:       vpnRange.String(),

		DnsServers:      strings.Join([]string{dnsIp.String(), "zerops"}, ", "),
		ServerAddress:   serverAddress,
		ServerPublicKey: response.GetVpn().GetServerPublicKey(),
	})
	configFile := filepath.Join(os.TempDir(), "zerops.conf")

	err = os.WriteFile(configFile, bf.Bytes(), 0777)
	if err != nil {
		return err
	}

	output, err := exec.Command("wireguard", "/installtunnelservice", configFile).Output()
	if err != nil {
		h.logger.Error(output)
		return err
	}

	return nil
}
