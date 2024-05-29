package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func envCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("env").
		Short(i18n.T(i18n.CmdDescEnv)).
		HelpFlag(i18n.T(i18n.CmdHelpEnv)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			cmdData.Stdout.PrintLines(
				printer.Style(styles.CobraSectionColor(), i18n.T(i18n.GlobalEnvVariables)),
				printer.Style(styles.CobraItemNameColor(), constants.CliLogFilePathEnvVar)+"\t"+i18n.T(i18n.CliLogFilePathEnvVar),
				printer.Style(styles.CobraItemNameColor(), constants.CliDataFilePathEnvVar)+"\t"+i18n.T(i18n.CliDataFilePathEnvVar),
				printer.Style(styles.CobraItemNameColor(), constants.CliTerminalMode)+"\t"+i18n.T(i18n.CliTerminalModeEnvVar),
				printer.EmptyLine,
				printer.Style(styles.CobraSectionColor(), i18n.T(i18n.CurrentlyUsedEnvVariables)),
			)

			body := uxBlock.NewTableBody()
			guestInfoPart(body)
			cmdData.UxBlocks.Table(body)

			return nil
		})
}

func guestInfoPart(tableBody *uxBlock.TableBody) {
	cliDataFilePath, _, err := constants.CliDataFilePath()
	if err != nil {
		cliDataFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoCliDataFilePath), cliDataFilePath)

	logFilePath, _, err := constants.LogFilePath()
	if err != nil {
		logFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoLogFilePath), logFilePath)

	wgConfigFilePath, _, err := constants.WgConfigFilePath()
	if err != nil {
		wgConfigFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoWgConfigFilePath), wgConfigFilePath)
}
