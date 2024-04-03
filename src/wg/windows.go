//go:build windows
// +build windows

package wg

import (
	"context"
	"io"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/output"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func CheckWgInstallation() error {
	_, err := exec.LookPath("wireguard")
	if err != nil {
		return errors.New(i18n.T(i18n.VpnWgQuickIsNotInstalledWindows))
	}

	return nil
}

func GenerateConfig(f io.Writer, privateKey wgtypes.Key, vpnSettings output.ProjectVpnItem) error {
	data, err := defaultTemplateData(privateKey, vpnSettings)
	if err != nil {
		return err
	}

	return template.Must(template.New("wg template").Parse(vpnTmpl)).Execute(f, data)
}

func UpCmd(ctx context.Context, filePath string) (err *exec.Cmd) {
	return exec.CommandContext(ctx, "wireguard", "/installtunnelservice", filePath)
}

func DownCmd(ctx context.Context, _, interfaceName string) (err *exec.Cmd) {
	return exec.CommandContext(ctx, "wireguard", "/uninstalltunnelservice", interfaceName)
}

var vpnTmpl = `
[Interface]
PrivateKey = {{.PrivateKey}}

Address = {{if .AssignedIpv4Address}}{{.AssignedIpv4Address}}/32{{end}}, {{.AssignedIpv6Address}}/128
DNS = {{.Ipv4NetworkGateway}}, zerops
### Alternative DNS
# PostUp = powershell -command "Add-DnsClientNrptRule -Namespace 'zerops' -NameServers '{{.Ipv4NetworkGateway}}'"
# PostDown = powershell -command "Get-DnsClientNrptRule | Where { $_.Namespace -match '.*zerops' } | Remove-DnsClientNrptRule -force"

[Peer]
PublicKey = {{.PublicKey}}

AllowedIPs = {{if .ProjectIpv4Network}}{{.ProjectIpv4Network}},{{end}} {{.ProjectIpv6Network}}, {{if .Ipv4Network}}{{.Ipv4Network}}, {{end}}{{.Ipv6Network}}

Endpoint = {{.ProjectIpv4SharedEndpoint}}

PersistentKeepalive = 5
`
