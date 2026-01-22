package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func vpnClearCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("clear").
		Short("Disconnects Zerops VPN, removes registered public key from the project via API and deletes locally stored wg key.").
		ScopeLevel(cmdBuilder.ScopeProject()).
		HelpFlag("Help for the 'vpn clear' command.").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			if vpnDownErr := disconnectVpn(ctx, cmdData.UxBlocks, false, false); vpnDownErr != nil {
				cmdData.UxBlocks.PrintWarningTextf("vpn down: %s", vpnDownErr)
			}

			if vpnKey, ok := cmdData.CliStorage.Data().ProjectVpnKeyRegistry[project.Id]; ok {
				wgKey, err := wgtypes.ParseKey(vpnKey.Key)
				if err != nil {
					return err
				}

				deleteResponse, err := cmdData.RestApiClient.DeleteProjectVpn(
					ctx,
					path.ProjectId{Id: project.Id},
					body.PostProjectVpn{PublicKey: types.String(wgKey.PublicKey().String())},
				)
				if err != nil {
					return err
				}
				if err := deleteResponse.Err(); err != nil {
					return err
				}

				cmdData.UxBlocks.PrintSuccessTextf("Removed registered public key from project via API: %s", wgKey.PublicKey().String())

				_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
					delete(data.ProjectVpnKeyRegistry, project.Id)
					cmdData.UxBlocks.PrintSuccessText("Deleted locally stored wg keys.")
					return data
				})
				if err != nil {
					return err
				}
			}

			return nil
		})
}
