package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func supportCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("support").
		Short(i18n.T(i18n.CmdDescSupport)).
		HelpFlag(i18n.T(i18n.CmdHelpSupport)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			cmdData.Stdout.PrintLines(
				printer.Style(styles.CobraSectionStyle(), i18n.T(i18n.Contact)),
				printer.Style(styles.CobraItemNameStyle(), "- E-mail")+":  team@zerops.io",
				printer.Style(styles.CobraItemNameStyle(), "- Discord")+": https://discord.com/invite/WDvCZ54",
				printer.EmptyLine,
				i18n.T(i18n.Documentation),
			)
			return nil
		})
}
