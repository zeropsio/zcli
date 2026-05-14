package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// EnvCmd prints supported env vars + a table of the currently resolved
// data/log/wg paths. Guest run.
type EnvCmd struct{}

func (*EnvCmd) Run(_ context.Context, d *kongCmd.CmdData) error {
	d.Stdout.PrintLines(
		printer.Style(styles.CobraSectionStyle(), i18n.T(i18n.GlobalEnvVariables)),
		printer.Style(styles.CobraItemNameStyle(), constants.CliLogFilePathEnvVar)+"\t"+i18n.T(i18n.CliLogFilePathEnvVar),
		printer.Style(styles.CobraItemNameStyle(), constants.CliDataFilePathEnvVar)+"\t"+i18n.T(i18n.CliDataFilePathEnvVar),
		printer.Style(styles.CobraItemNameStyle(), constants.CliTerminalMode)+"\t"+i18n.T(i18n.CliTerminalModeEnvVar),
		printer.EmptyLine,
		printer.Style(styles.CobraSectionStyle(), i18n.T(i18n.CurrentlyUsedEnvVariables)),
	)

	body := table.NewBody()
	guestInfoPart(body)
	d.Stdout.Println(table.Render(body))
	return nil
}
