//go:build linux

package wg

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/output"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func CheckWgInstallation() error {
	if _, err := exec.LookPath("wg-quick"); err != nil {
		return errors.New(i18n.T(i18n.VpnWgQuickIsNotInstalled))
	}
	// Debian does not have it by default anymore
	if _, err := exec.LookPath("resolvectl"); err != nil {
		return errors.New(i18n.T(i18n.VpnResolveCtlIsNotInstalled))
	}
	return nil
}

func GenerateConfig(f io.Writer, privateKey wgtypes.Key, vpnSettings output.ProjectVpnItem, mtu int) error {
	data, err := defaultTemplateData(privateKey, vpnSettings, mtu)
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
	interfaces, err := net.Interfaces()
	if err != nil {
		return false, errors.Wrap(err, "Can't resolve net interfaces")
	}
	_, found := gn.FindFirst(interfaces, func(in net.Interface) bool {
		fmt.Println(in.Name)
		return in.Name == constants.WgInterfaceName
	})
	return found, nil
}

var vpnTmpl = `
[Interface]
PrivateKey = {{.PrivateKey}}
MTU = {{.Mtu}}

Address = {{if .AssignedIpv4Address}}{{.AssignedIpv4Address}}/32{{end}}, {{.AssignedIpv6Address}}/128
PostUp = resolvectl dns %i {{.Ipv4NetworkGateway}}
PostUp = resolvectl domain %i zerops

[Peer]
PublicKey = {{.PublicKey}}

AllowedIPs = {{if .ProjectIpv4Network}}{{.ProjectIpv4Network}},{{end}} {{.ProjectIpv6Network}}, {{if .Ipv4Network}}{{.Ipv4Network}}, {{end}}{{.Ipv6Network}}

Endpoint = {{.ProjectIpv4SharedEndpoint}}

PersistentKeepalive = 5
`
