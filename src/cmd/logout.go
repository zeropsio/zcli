package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/loginSession"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/wg"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

func logoutCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("logout").
		Short(i18n.T(i18n.CmdDescLogout)).
		HelpFlag(i18n.T(i18n.CmdHelpLogout)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			uxBlocks := cmdData.UxBlocks
			data := cmdData.CliStorage.Data()
			envToken := cmdBuilder.EnvToken()
			hasStoredLogin := data.Token != "" || data.LoginSession.IsSet()

			// Personal tokens stay valid, only a browser login is revoked on the server.
			if data.LoginSession.IsSet() {
				revokeLoginSession(ctx, uxBlocks, cmdData.CliStorage, data.RegionData.Address)
			}

			_, err := cmdData.CliStorage.Clear()
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

			switch {
			case envToken != "":
				uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.LogoutEnvToken, constants.CliTokenEnvVar)))
			case !hasStoredLogin:
				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.LogoutNotLoggedIn)))
			default:
				uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.LogoutSuccess)))
			}
			return nil
		})
}

// revokeLoginSession goes through the refreshing source, because logging out with an expired access token would leave its refresh token valid.
func revokeLoginSession(ctx context.Context, uxBlocks uxBlock.UxBlocks, storage *cliStorage.Handler, host string) {
	if host == "" {
		host = constants.DefaultRegion
	}
	source := loginSession.New(storage, zeropsRestApiClient.NewClient(host))
	response, err := zeropsRestApiClient.NewTokenSourceClient(source, host).PostAuthLogout(ctx)
	if err == nil {
		_, err = response.Output()
	}
	if err != nil && !source.Expired() {
		uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.LogoutRevokeFailed, err.Error())))
	}
}
