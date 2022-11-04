//go:build windows

package vpn

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/zeropsio/zcli/src/i18n"
	vpnproxy "github.com/zeropsio/zcli/src/proto/vpnproxy"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const Template = `
[Interface]
PrivateKey = {{.ClientPrivateKey}}
Address = {{.ClientAddress}}
DNS = {{.DnsServers}}, zerops

[Peer]
PublicKey = {{.ServerPublicKey}}
AllowedIPs = {{.AllowedIPs}}
Endpoint = {{.ServerAddress}}
PersistentKeepalive = 25
`

func (h *Handler) setVpn(ctx context.Context, selectedVpnAddress net.IP, privateKey wgtypes.Key, mtu uint32, response *vpnproxy.StartVpnResponse) error {
	_, err := exec.LookPath("wireguard")
	if err != nil {
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
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
		ClientPrivateKey: privateKey.String(),
		ClientAddress:    clientIp.String(),
		AllowedIPs:       vpnRange.String(),

		DnsServers:      dnsIp.String(),
		ServerAddress:   net.JoinHostPort(selectedVpnAddress.String(), strconv.Itoa(int(response.GetVpn().GetPort()))),
		ServerPublicKey: response.GetVpn().GetServerPublicKey(),
	})
	configFile := filepath.Join(os.TempDir(), "zerops.conf")

	err = os.WriteFile(configFile, bf.Bytes(), 0777)
	if err != nil {
		return err
	}

	h.runCommands(ctx,
		makeCommand(
			"wireguard",
			commandWithErrorMessage(i18n.VpnStartTunnelConfigurationFailed),
			commandWithArgs("/installtunnelservice", configFile),
		),
		makeCommand(
			"netsh",
			commandWithErrorMessage(i18n.VpnStartTunnelConfigurationFailed),
			commandWithArgs("interface", "ipv4", "set", "subinterface", "zerops", "mtu="+strconv.Itoa(int(mtu))),
		),
	)

	time.Sleep(time.Second * 5)

	return nil
}
