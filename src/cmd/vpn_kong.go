package cmd

import (
	"context"
	"os"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/wg"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// VpnCmd groups the vpn subcommands.
type VpnCmd struct {
	Up     VpnUpCmd     `cmd:"" help:"Connect to the project VPN."`
	Down   VpnDownCmd   `cmd:"" help:"Disconnect from the project VPN."`
	Config VpnConfigCmd `cmd:"" help:"Generate VPN configuration file without connecting."`
	Clear  VpnClearCmd  `cmd:"" help:"Disconnect VPN, revoke the key from the project, delete local key."`
	Key    VpnKeyCmd    `cmd:"" help:"VPN keys commands group."`
}

// VpnUpCmd connects WireGuard to the project's VPN.
type VpnUpCmd struct {
	kongCmd.ProjectScope

	ProjectArg              string `arg:"" name:"project-id" optional:"" help:"Project ID."`
	Mtu                     int    `name:"mtu" default:"1420" help:"MTU for the WireGuard interface."`
	AutoDisconnect          bool   `name:"auto-disconnect" help:"Auto-disconnect any active VPN session."`
	SkipDnsSetup            bool   `name:"skip-dns-setup" help:"Skip DNS configuration."`
	SkipVpnTest             bool   `name:"skip-vpn-test" help:"Skip VPN connectivity test."`
	SkipCheckInstallation   bool   `name:"skip-check-installation" help:"Skip WireGuard installation check."`
}

func (c *VpnUpCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	if c.ProjectScope.ProjectID == "" && c.ProjectArg != "" {
		c.ProjectScope.ProjectID = c.ProjectArg
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	dnsSetup := !c.SkipDnsSetup
	vpnTest := !c.SkipVpnTest
	checkInstallation := !c.SkipCheckInstallation
	uxBlocks := d.UxBlocks

	vpnActive, err := wg.InterfaceExists()
	if err != nil {
		return err
	}
	if vpnActive {
		if c.AutoDisconnect {
			if err := disconnectVpn(ctx, uxBlocks, dnsSetup, checkInstallation); err != nil {
				return err
			}
		} else {
			confirmed, err := uxHelpers.YesNoPrompt(ctx, "VPN is active, do you want to disconnect?")
			if err != nil {
				return err
			}
			if !confirmed {
				uxBlocks.PrintInfoText("VPN is still connected")
				return nil
			}
			if err := disconnectVpn(ctx, uxBlocks, dnsSetup, checkInstallation); err != nil {
				return err
			}
		}
	}

	privateKey, err := vpnGetOrCreateKey(auth, project, uxBlocks)
	if err != nil {
		return err
	}
	publicKey := privateKey.PublicKey()

	postResp, err := auth.RestApiClient.PostProjectVpn(
		ctx,
		path.ProjectId{Id: project.Id},
		body.PostProjectVpn{PublicKey: types.String(publicKey.String())},
	)
	if err != nil {
		return err
	}
	vpnSettings, err := postResp.Output()
	if err != nil {
		return err
	}

	filePath, fileMode, err := constants.WgConfigFilePath()
	if err != nil {
		return err
	}
	f, err := file.Open(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := wg.GenerateConfig(f, privateKey, vpnSettings, c.Mtu, dnsSetup); err != nil {
		return err
	}
	uxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.VpnConfigSaved), filePath))

	if _, err := d.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		if data.ProjectVpnKeyRegistry == nil {
			data.ProjectVpnKeyRegistry = make(map[uuid.ProjectId]entity.VpnKey)
		}
		data.ProjectVpnKeyRegistry[project.Id] = entity.VpnKey{
			ProjectId: project.Id,
			Key:       privateKey.String(),
			CreatedAt: time.Now(),
		}
		return data
	}); err != nil {
		return err
	}

	if err := wg.CheckWgInstallation(checkInstallation, dnsSetup); err != nil {
		return err
	}
	if _, err := cmdRunner.Run(wg.UpCmd(ctx, filePath)); err != nil {
		return err
	}

	if vpnTest && dnsSetup {
		if isVpnUp(ctx, uxBlocks, 6) {
			uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.VpnUp)))
		} else {
			uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.VpnPingFailed)))
		}
	}
	return nil
}

// VpnDownCmd disconnects WireGuard. Guest-friendly (no API/scope).
type VpnDownCmd struct {
	SkipDnsSetup          bool `name:"skip-dns-setup" help:"Skip DNS configuration removal."`
	SkipCheckInstallation bool `name:"skip-check-installation" help:"Skip WireGuard installation check."`
}

func (c *VpnDownCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	if _, err := d.RequireAuth(); err != nil {
		return err
	}
	return disconnectVpn(ctx, d.UxBlocks, !c.SkipDnsSetup, !c.SkipCheckInstallation)
}

// VpnConfigCmd generates a WireGuard config file without connecting.
type VpnConfigCmd struct {
	kongCmd.ProjectScope

	ProjectArg   string `arg:"" name:"project-id" optional:"" help:"Project ID."`
	Mtu          int    `name:"mtu" default:"1420" help:"MTU for the WireGuard interface."`
	SkipDnsSetup bool   `name:"skip-dns-setup" help:"Skip DNS configuration."`
	Output       string `name:"output" help:"Output file path (use '-' for stdout, empty for default location)."`
}

func (c *VpnConfigCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	if c.ProjectScope.ProjectID == "" && c.ProjectArg != "" {
		c.ProjectScope.ProjectID = c.ProjectArg
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	dnsSetup := !c.SkipDnsSetup
	privateKey, err := vpnGetOrCreateKey(auth, project, d.UxBlocks)
	if err != nil {
		return err
	}
	publicKey := privateKey.PublicKey()

	postResp, err := auth.RestApiClient.PostProjectVpn(
		ctx,
		path.ProjectId{Id: project.Id},
		body.PostProjectVpn{PublicKey: types.String(publicKey.String())},
	)
	if err != nil {
		return err
	}
	vpnSettings, err := postResp.Output()
	if err != nil {
		return err
	}

	var f *os.File
	var filePath string
	var fileMode os.FileMode
	switch c.Output {
	case "-":
		f = os.Stdout
		filePath = "stdout"
	case "":
		filePath, fileMode, err = constants.WgConfigFilePath()
		if err != nil {
			return err
		}
		f, err = file.Open(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return err
		}
		defer f.Close()
	default:
		filePath = c.Output
		fileMode = 0o600
		f, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	if err := wg.GenerateConfig(f, privateKey, vpnSettings, c.Mtu, dnsSetup); err != nil {
		return err
	}
	if c.Output != "-" {
		d.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.VpnConfigSaved), filePath))
	}

	if _, err := d.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		if data.ProjectVpnKeyRegistry == nil {
			data.ProjectVpnKeyRegistry = make(map[uuid.ProjectId]entity.VpnKey)
		}
		data.ProjectVpnKeyRegistry[project.Id] = entity.VpnKey{
			ProjectId: project.Id,
			Key:       privateKey.String(),
			CreatedAt: time.Now(),
		}
		return data
	}); err != nil {
		return err
	}
	return nil
}

// VpnClearCmd disconnects VPN, removes the public key on the server, and
// deletes the local stored key.
type VpnClearCmd struct {
	kongCmd.ProjectScope
}

func (c *VpnClearCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	if vpnDownErr := disconnectVpn(ctx, d.UxBlocks, false, false); vpnDownErr != nil {
		d.UxBlocks.PrintWarningTextf("vpn down: %s", vpnDownErr)
	}

	vpnKey, ok := d.CliStorage.Data().ProjectVpnKeyRegistry[project.Id]
	if !ok {
		return nil
	}
	wgKey, err := wgtypes.ParseKey(vpnKey.Key)
	if err != nil {
		return err
	}
	deleteResp, err := auth.RestApiClient.DeleteProjectVpn(
		ctx,
		path.ProjectId{Id: project.Id},
		body.PostProjectVpn{PublicKey: types.String(wgKey.PublicKey().String())},
	)
	if err != nil {
		return err
	}
	if err := deleteResp.Err(); err != nil {
		return err
	}
	d.UxBlocks.PrintSuccessTextf("Removed registered public key from project via API: %s", wgKey.PublicKey().String())

	if _, err := d.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		delete(data.ProjectVpnKeyRegistry, project.Id)
		d.UxBlocks.PrintSuccessText("Deleted locally stored wg keys.")
		return data
	}); err != nil {
		return err
	}
	return nil
}

// vpnGetOrCreateKey returns the cached project key from Auth or generates
// a fresh WireGuard private key. The legacy cobra getter took the
// *LoggedUserCmdData; here we use Auth.VpnKeys directly.
func vpnGetOrCreateKey(auth *kongCmd.Auth, project entity.Project, ux *uxBlock.Blocks) (wgtypes.Key, error) {
	if existing, ok := auth.VpnKeys[project.Id]; ok {
		key, err := wgtypes.ParseKey(existing.Key)
		if err == nil {
			return key, nil
		}
		ux.PrintWarning(styles.WarningLine(i18n.T(i18n.VpnPrivateKeyCorrupted)))
	}
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, err
	}
	ux.PrintInfo(styles.InfoLine(i18n.T(i18n.VpnPrivateKeyCreated)))
	return key, nil
}
