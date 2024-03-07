package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func projectCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("project").
		Short(i18n.T(i18n.CmdProject)).
		HelpFlag(i18n.T(i18n.ProjectHelp)).
		AddChildrenCmd(projectListCmd()).
		AddChildrenCmd(projectStartCmd()).
		AddChildrenCmd(projectStopCmd()).
		AddChildrenCmd(projectDeleteCmd()).
		AddChildrenCmd(projectServiceImportCmd()).
		AddChildrenCmd(projectImportCmd())
}
