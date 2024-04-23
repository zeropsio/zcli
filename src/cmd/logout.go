package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/i18n"
    "github.com/zeropsio/zcli/src/uxBlock/styles"
)

var logout string

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

			_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
                return cliStorage.Data{}
			})
            if err != nil {
                return err
            }

            uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.LogoutVpnDisconnecting)))
            disconnectVpn(ctx, cmdData.UxBlocks) // TODO: ask - there is no need for any declaration - i can just call this function from anywhere?
            uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.LogoutSuccess)))

            return nil

        })
}