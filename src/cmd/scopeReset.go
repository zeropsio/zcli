package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func scopeResetCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("reset").
		Short(i18n.T(i18n.CmdDescScopeReset)).
		HelpFlag(i18n.T(i18n.CmdHelpScopeReset)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			if err := cmdBuilder.ProjectScopeReset(cmdData); err != nil {
				return err
			}
			if err := cmdBuilder.ServiceScopeReset(cmdData); err != nil {
				return err
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ScopeReset)))

			return nil
		})
}
