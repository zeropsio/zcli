package cmd

import (
	"context"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/nettools"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

const vpnCheckAddress = "logger.core.zerops"

func vpnUpCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("up").
		Short(i18n.T(i18n.CmdDescVpnUp)).
		ScopeLevel(scope.Project).
		Arg(scope.ProjectArgName, cmdBuilder.OptionalArg()).
		BoolFlag("auto-disconnect", false, i18n.T(i18n.VpnAutoDisconnectFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpVpnUp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			if isVpnUp(ctx) {
				if cmdData.Params.GetBool("auto-disconnect") {
					err := disconnectVpn(ctx, uxBlocks)
					if err != nil {
						return err
					}
				} else {
					confirmed, err := uxHelpers.YesNoPrompt(
						ctx,
						cmdData.UxBlocks,
						i18n.T(i18n.VpnDisconnectionPrompt),
					)
					if err != nil {
						return err
					}
					if !confirmed {
						return errors.New(i18n.T(i18n.VpnDisconnectionPromptNo))
					}

					err = disconnectVpn(ctx, uxBlocks)
					if err != nil {
						return err
					}
				}
			}

			privateKey, err := getOrCreatePrivateVpnKey(cmdData)
			if err != nil {
				return err
			}

			publicKey := privateKey.PublicKey()

			postProjectResponse, err := cmdData.RestApiClient.PostProjectVpn(
				ctx,
				path.ProjectId{Id: cmdData.Project.ID},
				body.PostProjectVpn{PublicKey: types.String(publicKey.String())},
			)
			if err != nil {
				return err
			}

			vpnSettings, err := postProjectResponse.Output()
			if err != nil {
				return err
			}

			filePath, fileMode, err := constants.WgConfigFilePath()
			if err != nil {
				return err
			}

			f, err := file.Open(filePath, os.O_RDWR|os.O_CREATE, fileMode)
			if err != nil {
				return err
			}
			err = func() error {
				defer f.Close()

				templ := template.Must(template.New("wg template").Parse(vpnTmpl))

				return templ.Execute(f, map[string]interface{}{
					"PrivateKey":                privateKey.String(),
					"PublicKey":                 vpnSettings.Project.PublicKey,
					"AssignedIpv4Address":       vpnSettings.Peer.Ipv4.AssignedIpAddress,
					"AssignedIpv6Address":       vpnSettings.Peer.Ipv6.AssignedIpAddress,
					"Ipv4NetworkGateway":        vpnSettings.Project.Ipv4.Network.Gateway,
					"ProjectIpv4Network":        vpnSettings.Project.Ipv4.Network.Network,
					"ProjectIpv6Network":        vpnSettings.Project.Ipv6.Network.Network,
					"Ipv4Network":               vpnSettings.Peer.Ipv4.Network.Network,
					"Ipv6Network":               vpnSettings.Peer.Ipv6.Network.Network,
					"ProjectIpv4SharedEndpoint": vpnSettings.Project.Ipv4.SharedEndpoint,
				})
			}()
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.VpnConfigSaved), filePath))

			_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				if data.VpnKeys == nil {
					data.VpnKeys = make(map[uuid.ProjectId]entity.VpnKey)
				}
				data.VpnKeys[cmdData.Project.ID] = entity.VpnKey{
					ProjectId: cmdData.Project.ID,
					Key:       privateKey.String(),
					CreatedAt: time.Now(),
				}

				return data
			})
			if err != nil {
				return err
			}

			_, err = exec.LookPath("wg-quick")
			if err != nil {
				return errors.New(i18n.T(i18n.VpnWgQuickIsNotInstalled))
			}

			c := exec.CommandContext(ctx, "wg-quick", "up", filePath)
			_, err = cmdRunner.Run(c)
			if err != nil {
				return err
			}

			if isVpnUp(ctx) {
				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.VpnUp)))
			} else {
				uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.VpnPingFailed)))
			}

			return nil
		})
}

func isVpnUp(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := nettools.Ping(ctx, vpnCheckAddress)
	return err == nil
}

func getOrCreatePrivateVpnKey(cmdData *cmdBuilder.LoggedUserCmdData) (wgtypes.Key, error) {
	projectId := cmdData.Project.ID

	if vpnKey, exists := cmdData.VpnKeys[projectId]; exists {
		wgKey, err := wgtypes.ParseKey(vpnKey.Key)
		if err == nil {
			return wgKey, nil
		}

		cmdData.UxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.VpnPrivateKeyCorrupted)))
	}

	vpnKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, err
	}

	cmdData.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.VpnPrivateKeyCreated)))

	return vpnKey, nil
}

var vpnTmpl = `
[Interface]
PrivateKey = {{.PrivateKey}}

Address = {{if .AssignedIpv4Address}}{{.AssignedIpv4Address}}/32{{end}}, {{.AssignedIpv6Address}}/128
DNS = {{.Ipv4NetworkGateway}}, zerops

[Peer]
PublicKey = {{.PublicKey}}

AllowedIPs = {{if .ProjectIpv4Network}}{{.ProjectIpv4Network}},{{end}} {{.ProjectIpv6Network}}, {{if .Ipv4Network}}{{.Ipv4Network}}, {{end}}{{.Ipv6Network}}

Endpoint = {{.ProjectIpv4SharedEndpoint}}

PersistentKeepalive = 5
`
