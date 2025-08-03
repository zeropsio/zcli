package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func projectCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("project").
		Short(i18n.T(i18n.CmdDescProject)).
		HelpFlag(i18n.T(i18n.CmdHelpProject)).
		AddChildrenCmd(projectCreateCmd()).
		AddChildrenCmd(projectListCmd()).
		AddChildrenCmd(projectDeleteCmd()).
		AddChildrenCmd(projectScopeCmd()).
		AddChildrenCmd(projectServiceImportCmd()).
		AddChildrenCmd(projectImportCmd()).
		AddChildrenCmd(projectEnvCmd())
}
