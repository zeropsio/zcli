package cmd

import (
	"context"
	_ "embed"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	repository2 "github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func statusInfoCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("info").
		Short(i18n.T(i18n.CmdStatusInfo)).
		HelpFlag(i18n.T(i18n.StatusInfoHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			body := &uxBlock.TableBody{}

			cliDataFilePath, err := constants.CliDataFilePath()
			if err != nil {
				cliDataFilePath = err.Error()
			}
			body.AddStringsRow(i18n.T(i18n.StatusInfoCliDataFilePath), cliDataFilePath)

			logFilePath, err := constants.LogFilePath()
			if err != nil {
				logFilePath = err.Error()
			}
			body.AddStringsRow(i18n.T(i18n.StatusInfoLogFilePath), logFilePath)

			if cmdData.CliStorage.Data().ScopeProjectId.Filled() {
				// project scope is set
				projectId, _ := cmdData.CliStorage.Data().ScopeProjectId.Get()
				project, err := repository2.GetProjectById(ctx, cmdData.RestApiClient, projectId)
				if err != nil {
					if errorsx.IsUserError(err) {
						cmdData.UxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.ScopedProjectNotFound)))
					}

					return err
				}

				body.AddStringsRow(i18n.T(i18n.ScopedProject), project.Name.String())
			}

			cmdData.UxBlocks.Table(body)

			return nil
		})
}
