package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/wg"
)

func logoutCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("logout").
		Short(i18n.T(i18n.CmdDescLogout)).
		HelpFlag(i18n.T(i18n.CmdHelpLogout)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			_, err := cmdData.RestApiClient.PostAuthLogout(ctx)
			if err != nil {
				return err
			}

			_, err = cmdData.CliStorage.Clear()
			if err != nil {
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
		})
}
