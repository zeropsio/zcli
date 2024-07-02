package cmd

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/wg"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func vpnUpCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("up").
		Alias("u", "start").
		Short(i18n.T(i18n.CmdDescVpnUp)).
		ScopeLevel(scope.Project).
		Arg(scope.ProjectArgName, cmdBuilder.OptionalArg()).
		BoolFlag("auto-disconnect", false, i18n.T(i18n.VpnAutoDisconnectFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpVpnUp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			if isVpnUp(uxBlocks) {
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

			f, err := file.Open(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
			if err != nil {
				return err
			}
			defer f.Close()

			err = wg.GenerateConfig(f, privateKey, vpnSettings)
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

			err = wg.CheckWgInstallation()
			if err != nil {
				return err
			}

			c := wg.UpCmd(ctx, filePath)
			_, err = cmdRunner.Run(c)
			if err != nil {
				return err
			}

			// wait for the vpn to be up
			if isVpnUp(uxBlocks) {
				uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.VpnUp)))
			} else {
				uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.VpnPingFailed)))
			}

			return nil
		})
}

// errNoSuchInterface copied from 'net' package, because it's private :)
var errNoSuchInterface = errors.New("no such network interface")

func isVpnUp(uxBlocks uxBlock.UxBlocks) bool {
	_, err := net.InterfaceByName(constants.WgInterfaceName)
	opError := &net.OpError{}
	// cannot use errors.Is(), because std error package does not implement Is() interface, so we have this abomination
	if errors.As(err, &opError) && opError.Err.Error() != errNoSuchInterface.Error() {
		uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.WarnVpnInterface, opError.Err.Error())))
	}
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
