package cmd

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/nettools"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/wg"
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
		ScopeLevel(cmdBuilder.ScopeProject()).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		IntFlag("mtu", 1420, i18n.T(i18n.VpnMtuFlag)).
		BoolFlag("auto-disconnect", false, i18n.T(i18n.VpnAutoDisconnectFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpVpnUp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			if isVpnUp(ctx, uxBlocks, 1) {
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

			privateKey, err := getOrCreatePrivateVpnKey(project, cmdData)
			if err != nil {
				return err
			}

			publicKey := privateKey.PublicKey()

			postProjectResponse, err := cmdData.RestApiClient.PostProjectVpn(
				ctx,
				path.ProjectId{Id: project.ID},
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

			err = wg.GenerateConfig(f, privateKey, vpnSettings, cmdData.Params.GetInt("mtu"))
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.VpnConfigSaved), filePath))

			_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				if data.VpnKeys == nil {
					data.VpnKeys = make(map[uuid.ProjectId]entity.VpnKey)
				}
				data.VpnKeys[project.ID] = entity.VpnKey{
					ProjectId: project.ID,
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
			if isVpnUp(ctx, uxBlocks, 6) {
				uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.VpnUp)))
			} else {
				uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.VpnPingFailed)))
			}

			return nil
		})
}

func isVpnUp(ctx context.Context, uxBlocks *uxBlock.Blocks, attempts int) bool {
	p := []uxHelpers.Process{
		{
			F: func(ctx context.Context, _ *uxHelpers.Process) error {
				for i := 0; i < attempts; i++ {
					err := nettools.Ping(ctx, vpnCheckAddress)
					if err == nil {
						return nil
					}

					time.Sleep(time.Millisecond * 500)
				}
				return errors.New(i18n.T(i18n.VpnPingFailed))
			},
			RunningMessage:      i18n.T(i18n.VpnCheckingConnection),
			ErrorMessageMessage: "",
			SuccessMessage:      "",
		},
	}

	err := uxHelpers.ProcessCheckWithSpinner(ctx, uxBlocks, p)

	return err == nil
}

func getOrCreatePrivateVpnKey(project entity.Project, cmdData *cmdBuilder.LoggedUserCmdData) (wgtypes.Key, error) {
	projectId := project.ID

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
