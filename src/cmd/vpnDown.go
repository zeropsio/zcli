package cmd

import (
	"context"
	"os"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/wg"
)

func vpnDownCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("down").
		Short(i18n.T(i18n.CmdDescVpnDown)).
		HelpFlag(i18n.T(i18n.CmdHelpVpnDown)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			return disconnectVpn(ctx, cmdData.UxBlocks)
		})
}

func disconnectVpn(ctx context.Context, uxBlocks uxBlock.UxBlocks) error {
	err := wg.CheckWgInstallation()
	if err != nil {
		return err
	}

	filePath, fileMode, err := constants.WgConfigFilePath()
	if err != nil {
		return err
	}

	// create empty file if not exists, only thing wg-quick needs is a proper file name
	f, err := file.Open(filePath, os.O_RDWR|os.O_CREATE, fileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	c := wg.DownCmd(ctx, filePath, constants.WgInterfaceName)
	_, err = cmdRunner.Run(c)
	if err != nil {
		return err
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.VpnDown)))

	return nil
}
