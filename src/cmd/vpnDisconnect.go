package cmd

import (
	"context"
	"os/exec"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func vpnDisconnectCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("disconnect").
		Short(i18n.T(i18n.CmdVpnDisconnect)).
		HelpFlag(i18n.T(i18n.VpnDisconnectHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			// TODO - janhajek check if vpn is connected
			// TODO - janhajek get somehow a meaningful output
			// TODO - janhajek check if wg-quick is installed
			// TODO - janhajek a configurable path to wg-quick
			c := exec.CommandContext(ctx, "wg-quick", "down", "zerops")
			_, err := cmdRunner.Run(c)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.VpnDisconnected)))

			return nil
		})
}
