package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func scopeServiceCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("service").
		Short("Sets the scope for a service. Commands that require a service ID will use the selected one.").
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		HelpFlag("Help for the scope service command.").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			if !cmdData.Profile.HasRepo() {
				return errors.New("`zcli scope service` only works inside a repo with a profile; run `zcli profile init` here first")
			}

			if serviceId, serviceSet := cmdData.Profile.ScopeService(); serviceSet {
				service, err := repository.GetServiceById(ctx, cmdData.RestApiClient, serviceId)
				if err != nil {
					if errorsx.Is(err, errorsx.ErrorCode(errorsx.CodeServiceStackNotFound)) {
						cmdData.UxBlocks.PrintWarning(styles.WarningLine(err.Error()))
					} else {
						return err
					}
				} else {
					cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Previously scoped service", service.Name))
				}
			}

			// Resolve the project first so the service selector has the right project context. This
			// mirrors the implicit lookup ScopeService(...) uses at command-execution time.
			projectId, projectSet := cmdData.Profile.ScopeProject()
			var project entity.Project
			var err error
			if projectSet {
				project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
				if err != nil {
					return err
				}
			} else {
				selectedProject, err := uxHelpers.PrintProjectSelector(ctx, cmdData.RestApiClient)
				if err != nil {
					return err
				}
				project, err = selectedProject.Expect("project is null")
				if err != nil {
					return err
				}
			}

			var service entity.Service
			if len(cmdData.Args[cmdBuilder.ServiceArgName]) > 0 {
				arg := cmdData.Args[cmdBuilder.ServiceArgName][0]
				if len(arg) == 22 {
					service, err = repository.GetServiceById(ctx, cmdData.RestApiClient, entity.ServiceId(arg))
				} else {
					service, err = repository.GetServiceByIdOrName(ctx, cmdData.RestApiClient, project.Id, arg)
				}
				if err != nil {
					return err
				}
			} else {
				selected, err := uxHelpers.PrintServiceSelector(ctx, cmdData.RestApiClient, project)
				if err != nil {
					return err
				}
				service, err = selected.Expect("service is null")
				if err != nil {
					return err
				}
			}

			if err := cmdData.Profile.SetScopeService(optional.New(service.Id)); err != nil {
				return err
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected service", service.Name))

			return nil
		})
}
