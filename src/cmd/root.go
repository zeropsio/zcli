package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
)

func ExecuteCmd() error {
	builder := cmdBuilder.NewCmdBuilder()

	builder.AddCommand(loginCmd())
	builder.AddCommand(versionCmd())
	builder.AddCommand(scopeCmd())
	builder.AddCommand(projectCmd())
	builder.AddCommand(serviceCmd())
	builder.AddCommand(statusCmd())

	return builder.CreateAndExecuteRootCobraCmd()
}
