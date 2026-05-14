package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// SupportCmd prints contact info + a link to the docs. Guest-friendly.
type SupportCmd struct{}

func (*SupportCmd) Run(_ context.Context, d *kongCmd.CmdData) error {
	d.Stdout.PrintLines(
		printer.Style(styles.CobraSectionStyle(), i18n.T(i18n.Contact)),
		printer.Style(styles.CobraItemNameStyle(), "- E-mail")+":  team@zerops.io",
		printer.Style(styles.CobraItemNameStyle(), "- Discord")+": https://discord.com/invite/WDvCZ54",
		printer.EmptyLine,
		i18n.T(i18n.Documentation),
	)
	return nil
}
