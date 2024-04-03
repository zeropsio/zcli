//go:build darwin
// +build darwin

package wg

import (
	"context"
	"io"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/output"
)

func CheckWgInstallation() error {
	_, err := exec.LookPath("wg-quick")
	if err != nil {
		return errors.New(i18n.T(i18n.VpnWgQuickIsNotInstalled))
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
	return exec.CommandContext(ctx, "wg-quick", "up", filePath)
}

func DownCmd(ctx context.Context, filePath, _ string) (err *exec.Cmd) {
	return exec.CommandContext(ctx, "wg-quick", "down", filePath)
}

var vpnTmpl = `
[Interface]
PrivateKey = {{.PrivateKey}}

Address = {{if .AssignedIpv4Address}}{{.AssignedIpv4Address}}/32{{end}}, {{.AssignedIpv6Address}}/128
PostUp = mkdir -p /etc/resolver 
PostUp = echo "nameserver {{.Ipv4NetworkGateway}}" > /etc/resolver/zerops 
PostUp = echo "domain zerops" >> /etc/resolver/zerops 
PostDown = rm /etc/resolver/zerops 

[Peer]
PublicKey = {{.PublicKey}}

AllowedIPs = {{if .ProjectIpv4Network}}{{.ProjectIpv4Network}},{{end}} {{.ProjectIpv6Network}}, {{if .Ipv4Network}}{{.Ipv4Network}}, {{end}}{{.Ipv6Network}}

Endpoint = {{.ProjectIpv4SharedEndpoint}}

PersistentKeepalive = 5
`
