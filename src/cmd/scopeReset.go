package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func scopeResetCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("reset").
		Short(i18n.T(i18n.CmdScopeReset)).
		HelpFlag(i18n.T(i18n.ScopeResetHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			err := scope.ProjectScopeReset(cmdData)
			if err != nil {
				return err
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ScopeReset)))

			return nil
		})
}
