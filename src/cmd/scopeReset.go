package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func scopeResetCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("reset").
		Short(i18n.T(i18n.CmdScopeReset)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			_, err := cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				data.ScopeProjectId = uuid.ProjectIdNull{}
				return data
			})
			if err != nil {
				return err
			}

			cmdData.UxBlocks.PrintInfoLine(i18n.T(i18n.ScopeReset))

			return nil
		})
}
