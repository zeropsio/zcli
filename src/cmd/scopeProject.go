package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func scopeProjectCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("project").
		Short(i18n.T(i18n.CmdDescScopeProject)).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpScopeProject)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			if !cmdData.Profile.HasRepo() {
				return errors.New("`zcli scope project` only works inside a repo with a profile; run `zcli profile init` here first")
			}

			if projectId, projectSet := cmdData.Profile.ScopeProject(); projectSet {
				project, err := repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
				if err != nil {
					if errorsx.Is(err, errorsx.ErrorCode(errorsx.CodeProjectNotFound)) {
						cmdData.UxBlocks.PrintWarning(styles.WarningLine(err.Error()))
					} else {
						return err
					}
				} else {
					cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.PreviouslyScopedProject), project.Name))
				}
			}

			infoText := i18n.SelectedProject
			var project entity.Project
			var err error

			if len(cmdData.Args) > 0 {
				project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, entity.ProjectId(cmdData.Args["project-id"][0]))
				if err != nil {
					return err
				}
			} else {
				// interactive selector of a project
				p, err := uxHelpers.PrintProjectSelector(ctx, cmdData.RestApiClient)
				if err != nil {
					return err
				}
				project, err = p.Expect("project is null")
				if err != nil {
					return err
				}
			}

			if err := cmdData.Profile.SetScopeProject(optional.New(project.Id)); err != nil {
				return err
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(infoText), project.Name))

			return nil
		})
}
