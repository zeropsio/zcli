//go:build windows
// +build windows

package wg

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/output"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// To install wireguard tunnel in windows wireguard.exe has to be run with elevated permissions.
// Only (simple) way I found to achieve this is to run Start-Process cmdlet with param '-Verb RunAS'
// https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.management/start-process?view=powershell-7.4

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

func UpCmd(ctx context.Context, filePath string) (err *cmdRunner.ExecCmd) {
	return cmdRunner.CommandContext(ctx,
		"powershell",
		"-Command",
		"Start-Process", "wireguard",
		"-Verb", "RunAs",
		"-ArgumentList "+formatArgumentList("/installtunnelservice", filePath),
	).
		SetBefore(beforeUp(filePath))
}

// beforeUp this function tries to remove previous zerops.conf from usual wireguard configuration
// dir (at %ProgramFiles%\WireGuard\Data\Configurations) and copy a newly generated one.
// It fails with error = nil because it's only for windows wireguard GUI.
func beforeUp(zeropsConfPath string) cmdRunner.Func {
	return func(_ context.Context) error {
		programFiles, set := os.LookupEnv("ProgramFiles")
		if !set {
			return nil
		}

		wgConfigDir := filepath.Join(programFiles, "WireGuard", "Data", "Configurations")
		stat, err := os.Stat(wgConfigDir)
		if err != nil {
			//nolint:nilerr
			return nil
		}
		if !stat.IsDir() {
			return nil
		}

		wgConfFile := filepath.Join(wgConfigDir, constants.WgConfigFile)
		// remove previous zerops.conf encrypted by wireguard.exe thus ending with .dpapi
		// https://git.zx2c4.com/wireguard-windows/about/docs/enterprise.md
		_ = os.Remove(wgConfFile + ".dpapi")

		wgConf, err := os.OpenFile(filepath.Join(wgConfigDir, constants.WgConfigFile), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			//nolint:nilerr
			return nil
		}
		defer wgConf.Close()

		zeropsConf, err := os.OpenFile(zeropsConfPath, os.O_RDONLY, 0666)
		if err != nil {
			_ = os.Remove(wgConfFile)
			//nolint:nilerr
			return nil
		}
		defer zeropsConf.Close()

		_, err = io.Copy(wgConf, zeropsConf)
		if err != nil {
			_ = os.Remove(wgConfFile)
			//nolint:nilerr
			return nil
		}

		return nil
	}
}

func DownCmd(ctx context.Context, _, interfaceName string) (err *cmdRunner.ExecCmd) {
	return cmdRunner.CommandContext(ctx,
		"powershell",
		"-Command",
		"Start-Process", "wireguard",
		"-Verb", "RunAs",
		"-ArgumentList "+formatArgumentList("/uninstalltunnelservice", interfaceName),
	)
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

func formatArgumentList(args ...string) string {
	for i, a := range args {
		args[i] = quote(a)
	}
	return strings.Join(args, ", ")
}

func quote(in string) string {
	return `"` + in + `"`
}
