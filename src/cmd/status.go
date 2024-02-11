package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func statusCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("status").
		Short(i18n.T(i18n.CmdStatus)).
		Short(i18n.T(i18n.StatusHelp)).
		AddChildrenCmd(statusShowDebugLogsCmd()).
		AddChildrenCmd(statusInfoCmd())
}
