package cmdBuilder

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type project struct {
	commonDependency
}

const ProjectArgName = "projectId"

func (p *project) AddCommandFlags(cmd *Cmd) {
	// TODO - janhajek translation
	cmd.StringFlag(ProjectArgName, "", "Project id")
}

func (p *project) LoadSelectedScope(ctx context.Context, cmd *Cmd, cmdData *LoggedUserCmdData) error {
	infoText := i18n.SelectedProject
	var project *entity.Project
	var err error

	// project scope is set
	if cmdData.CliStorage.Data().ScopeProjectId.Filled() {
		projectId, _ := cmdData.CliStorage.Data().ScopeProjectId.Get()

		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
		if err != nil {
			if errorsx.IsUserError(err) {
				cmdData.UxBlocks.PrintWarningLine(i18n.T(i18n.ScopedProjectNotFound))
			}

			return err
		}

		infoText = i18n.ScopedProject
	}

	if projectId, exists := cmdData.Args[ProjectArgName]; exists {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, uuid.ProjectId(projectId[0]))
		if err != nil {
			return err
		}

		infoText = i18n.SelectedProject
	}

	// service id is passed as a flag
	if projectId := cmdData.Params.GetString(ProjectArgName); projectId != "" {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, uuid.ProjectId(projectId))
		if err != nil {
			return err
		}

		infoText = i18n.SelectedProject
	}

	if project == nil {
		// interactive selector of a project
		project, err = uxHelpers.PrintProjectSelector(ctx, cmdData.UxBlocks, cmdData.RestApiClient)
		if err != nil {
			return err
		}
	}

	cmdData.Project = project
	cmdData.UxBlocks.PrintInfoLine(i18n.T(infoText, cmdData.Project.Name.String()))

	return nil
}
