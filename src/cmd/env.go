package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

var env string

func envCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("env").
		Short(i18n.T(i18n.CmdDescEnv)).
		HelpFlag(i18n.T(i18n.CmdHelpEnv)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {

			fmt.Println(styles.CobraSectionColor().SetString("Global Env Variables:").String() + `
` + styles.CobraItemNameColor().SetString(constants.CliLogFilePathEnvVar).String() + `     ` + i18n.T(i18n.CliLogFilePathEnvVar) + `
` + styles.CobraItemNameColor().SetString(constants.CliDataFilePathEnvVar).String() + `    ` + i18n.T(i18n.CliDataFilePathEnvVar) + `
` + styles.CobraItemNameColor().SetString(constants.CliTerminalMode).String() + `     ` + i18n.T(i18n.CliTerminalModeEnvVar))

fmt.Println(styles.CobraSectionColor().SetString(`
Curently used variables:`).String())

		    body := &uxBlock.TableBody{}
            guestInfoPart(body)
            cmdData.UxBlocks.Table(body)

			return nil
		})
}