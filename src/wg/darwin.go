//go:build darwin

package wg

import (
	"context"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/output"
)

const wgRunDir = "/var/run/wireguard/"

func CheckWgInstallation(checkInstallation, _ bool) error {
	if !checkInstallation {
		return nil
	}
	_, err := exec.LookPath("wg-quick")
	if err != nil {
		return errors.New(i18n.T(i18n.VpnWgQuickIsNotInstalled))
	}

	return nil
}

func GenerateConfig(f io.Writer, privateKey wgtypes.Key, vpnSettings output.ProjectVpnItem, mtu int, dnsSetup bool) error {
	data, err := defaultTemplateData(privateKey, vpnSettings, mtu, dnsSetup)
	if err != nil {
		return err
	}

	return template.Must(template.New("wg template").Parse(vpnTmpl)).Execute(f, data)
}

func UpCmd(ctx context.Context, filePath string) (err *cmdRunner.ExecCmd) {
	return cmdRunner.CommandContext(ctx, "wg-quick", "up", filePath)
}

func DownCmd(ctx context.Context, filePath, _ string) (err *cmdRunner.ExecCmd) {
	return cmdRunner.CommandContext(ctx, "wg-quick", "down", filePath)
}

func InterfaceExists() (bool, error) {
	items, err := os.ReadDir(wgRunDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "Cannot read /var/run/wireguard/ dir")
	}
	for _, i := range items {
		if i.Type()&os.ModeSocket != 0 {
			continue
		}
		if i.Name() == constants.WgInterfaceName+".name" {
			return true, nil
		}
	}
	return false, nil
}

var vpnTmpl = `
[Interface]
PrivateKey = {{.PrivateKey}}
MTU = {{.Mtu}}

Address = {{if .AssignedIpv4Address}}{{.AssignedIpv4Address}}/32{{end}}, {{.AssignedIpv6Address}}/128
{{if .DnsSetup -}}
PostUp = mkdir -p /etc/resolver 
PostUp = echo "nameserver {{.Ipv4NetworkGateway}}" > /etc/resolver/zerops 
PostUp = echo "domain zerops" >> /etc/resolver/zerops
PostUp = echo "search zerops" >> /etc/resolver/zerops
PostDown = rm /etc/resolver/zerops 
{{end}}

[Peer]
PublicKey = {{.PublicKey}}

AllowedIPs = {{if .ProjectIpv4Network}}{{.ProjectIpv4Network}},{{end}} {{.ProjectIpv6Network}}, {{if .Ipv4Network}}{{.Ipv4Network}}, {{end}}{{.Ipv6Network}}

Endpoint = {{.ProjectIpv4SharedEndpoint}}

PersistentKeepalive = 5
`
