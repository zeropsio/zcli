package scope

import (
	"context"

	"github.com/zeropsio/zcli/src/cliStorage"
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

type project struct {
}

const ProjectArgName = "projectId"

func (p *project) GetParent() cmdBuilder.ScopeLevel {
	return nil
}

func (p *project) AddCommandFlags(cmd *cmdBuilder.Cmd) {
	cmd.StringFlag(ProjectArgName, "", i18n.T(i18n.ProjectIdFlag))
}

func (p *project) LoadSelectedScope(ctx context.Context, cmd *cmdBuilder.Cmd, cmdData *cmdBuilder.LoggedUserCmdData) error {
	infoText := i18n.SelectedProject
	var project *entity.Project
	var err error

	// project scope is set
	if cmdData.CliStorage.Data().ScopeProjectId.Filled() {
		projectId, _ := cmdData.CliStorage.Data().ScopeProjectId.Get()

		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
		if err != nil {
			if errorsx.Check(err, errorsx.CheckErrorCode(errorCode.ProjectNotFound)) {
				err := ProjectScopeReset(cmdData)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			infoText = i18n.ScopedProject
		}
	}

	if projectId, exists := cmdData.Args[ProjectArgName]; exists {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, uuid.ProjectId(projectId[0]))
		if err != nil {
			return errorsx.Convert(
				err,
				errorsx.ConvertInvalidUserInput("id", i18n.T(i18n.ErrorInvalidProjectId)),
			)
		}

		infoText = i18n.SelectedProject
	}

	// service id is passed as a flag
	if projectId := cmdData.Params.GetString(ProjectArgName); projectId != "" {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, uuid.ProjectId(projectId))
		if err != nil {
			return errorsx.Convert(
				err,
				errorsx.ConvertInvalidUserInput("id", i18n.T(i18n.ErrorInvalidProjectId)),
			)
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
	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(infoText), cmdData.Project.Name.String()))

	return nil
}

func ProjectScopeReset(cmdData *cmdBuilder.LoggedUserCmdData) error {
	_, err := cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.ScopeProjectId = uuid.ProjectIdNull{}
		return data
	})
	if err != nil {
		return err
	}

	return nil
}
