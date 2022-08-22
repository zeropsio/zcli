//go:build windows
// +build windows

package vpn

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	vpnproxy "github.com/zerops-io/zcli/src/proto/vpnproxy"
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
PersistentKeepalive = 25
`

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, mtu uint32, response *vpnproxy.StartVpnResponse) error {
	_, err := exec.LookPath("wireguard")
	if err != nil {
		return err
	}

	clientIp := vpnproxy.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := vpnproxy.FromProtoIPRange(response.GetVpn().GetVpnIpRange())
	dnsIp := vpnproxy.FromProtoIP(response.GetVpn().GetDnsIp())
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
		ServerAddress:   selectedVpnAddress,
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
