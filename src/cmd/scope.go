package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func scopeCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("scope").
		Short(i18n.T(i18n.CmdDescScope)).
		HelpFlag(i18n.T(i18n.CmdHelpScope)).
		AddChildrenCmd(scopeProjectCmd()).
		AddChildrenCmd(scopeResetCmd())
}
