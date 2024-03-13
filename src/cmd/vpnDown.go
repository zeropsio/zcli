package cmd

import (
	"context"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func vpnDownCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("down").
		Short(i18n.T(i18n.CmdVpnDown)).
		HelpFlag(i18n.T(i18n.VpnDownHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			return disconnectVpn(ctx, cmdData.UxBlocks)
		})
}

func disconnectVpn(ctx context.Context, uxBlocks uxBlock.UxBlocks) error {
	_, err := exec.LookPath("wg-quick")
	if err != nil {
		return errors.New(i18n.T(i18n.VpnWgQuickIsNotInstalled))
	}

	filePath, err := constants.WgConfigFilePath()
	if err != nil {
		return err
	}

	// create empty file if not exists, only thing wg-quick needs is a proper file name
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	c := exec.CommandContext(ctx, "wg-quick", "down", filePath)
	_, err = cmdRunner.Run(c)
	if err != nil {
		return err
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.VpnDown)))

	return nil
}
