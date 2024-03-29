package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func scopeProjectCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("project").
		Short(i18n.T(i18n.CmdDescScopeProject)).
		Arg(scope.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpScopeProject)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			projectId, projectSet := cmdData.CliStorage.Data().ScopeProjectId.Get()
			if projectSet {
				project, err := repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
				if err != nil {
					if errorsx.Is(err, errorsx.ErrorCode(errorCode.ProjectNotFound)) {
						cmdData.UxBlocks.PrintWarning(styles.WarningLine(err.Error()))
					} else {
						return err
					}
				} else {
					cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.PreviouslyScopedProject), project.Name.String()))
				}
			}

			infoText := i18n.SelectedProject
			var project *entity.Project
			var err error

			if len(cmdData.Args) > 0 {
				project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, uuid.ProjectId(cmdData.Args["projectId"][0]))
				if err != nil {
					return err
				}
			} else {
				// interactive selector of a project
				project, err = uxHelpers.PrintProjectSelector(ctx, cmdData.UxBlocks, cmdData.RestApiClient)
				if err != nil {
					return err
				}
			}

			_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				data.ScopeProjectId = project.ID.ProjectIdNull()
				return data
			})
			if err != nil {
				return err
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(infoText), project.Name.String()))

			return nil
		})
}
