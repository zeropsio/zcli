package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/wg"
)

// LogoutCmd revokes the active token on the server, clears local storage,
// and disconnects the VPN if it's up. Requires auth.
type LogoutCmd struct{}

func (*LogoutCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	uxBlocks := d.UxBlocks

	if _, err := auth.RestApiClient.PostAuthLogout(ctx); err != nil {
		return err
	}
	if _, err := d.CliStorage.Clear(); err != nil {
		return err
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.LogoutVpnDisconnecting)))
	vpnActive, err := wg.InterfaceExists()
	if err != nil {
		return err
	}
	if vpnActive {
		_ = disconnectVpn(ctx, uxBlocks, false, false)
	}
	uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.LogoutSuccess)))

	return nil
}
