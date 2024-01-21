package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
)

func ExecuteCmd() error {
	cmdBuilder := cmdBuilder.NewCmdBuilder()

	cmdBuilder.AddCommand(loginCmd())
	cmdBuilder.AddCommand(versionCmd())
	cmdBuilder.AddCommand(scopeCmd())
	cmdBuilder.AddCommand(projectCmd())
	cmdBuilder.AddCommand(serviceCmd())
	cmdBuilder.AddCommand(statusCmd())
	cmdBuilder.AddCommand(bucketCmd())

	return cmdBuilder.CreateAndExecuteRootCobraCmd()
}
