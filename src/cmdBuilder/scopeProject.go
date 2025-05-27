package cmdBuilder

import (
	"context"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type ProjectOption gn.Option[projectScope]

// WithCreateNewProject allows 'create new project' option in selector
func WithCreateNewProject() ProjectOption {
	return func(s *projectScope) {
		s.createNew = true
	}
}

type projectScope struct {
	createNew bool
}

func ScopeProject(opts ...ProjectOption) ScopeLevel {
	return gn.ApplyOptions(opts...)
}

const ProjectArgName = "project-id"

func (p *projectScope) AddCommandFlags(cmd *Cmd) {
	cmd.StringFlag(ProjectArgName, "", i18n.T(i18n.ProjectIdFlag), ShortHand("P"))
}

func (p *projectScope) LoadSelectedScope(ctx context.Context, _ *Cmd, cmdData *LoggedUserCmdData) error {
	var project entity.Project
	var err error

	// service scope is set - use project from it
	if service, filled := cmdData.Service.Get(); filled {
		project, err := repository.GetProjectById(ctx, cmdData.RestApiClient, service.ProjectID)
		if err == nil {
			cmdData.Project = optional.New(project)
			return nil
		}
		cmdData.Project = optional.New(project)
	}

	// project scope is set
	if cmdData.CliStorage.Data().ScopeProjectId.Filled() {
		projectId, _ := cmdData.CliStorage.Data().ScopeProjectId.Get()

		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
		if err != nil {
			if errorsx.Is(err, errorsx.ErrorCode(errorCode.ProjectNotFound)) {
				err := ProjectScopeReset(cmdData)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		cmdData.Project = optional.New(project)
	}

	// project id is passed as a flag
	if projectId := cmdData.Params.GetString(ProjectArgName); projectId != "" {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, uuid.ProjectId(projectId))
		if err != nil {
			return errorsx.Convert(
				err,
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ apiError.Error, metaItemTyped map[string]interface{}) string {
							return i18n.T(i18n.ErrorInvalidProjectId, projectId, metaItemTyped["message"])
						},
					),
				),
			)
		}
		cmdData.Project = optional.New(project)
	}

	if projectId, exists := cmdData.Args[ProjectArgName]; exists && !cmdData.Project.Filled() {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, uuid.ProjectId(projectId[0]))
		if err != nil {
			return errorsx.Convert(
				err,
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ apiError.Error, metaItemTyped map[string]interface{}) string {
							return i18n.T(i18n.ErrorInvalidProjectId, projectId, metaItemTyped["message"])
						},
					),
				),
			)
		}
		cmdData.Project = optional.New(project)
	}

	if !cmdData.Project.Filled() {
		// interactive selector of a project
		selectedProject, err := uxHelpers.PrintProjectSelector(
			ctx,
			cmdData.RestApiClient,
			uxHelpers.WithCreateNewProject(p.createNew),
		)
		if err != nil {
			return err
		}

		if selectedProject.Filled() {
			project = selectedProject.Some()
		} else if terminal.IsTerminal() {
			project, err = createNewProject(ctx, cmdData)
			if err != nil {
				return err
			}
		}

		cmdData.Project = optional.New(project)
	}

	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.SelectedProject), project.Name.String()))

	return nil
}

func createNewProject(ctx context.Context, cmdData *LoggedUserCmdData) (entity.Project, error) {
	var err error
	selectedOrg, err := uxHelpers.PrintOrgSelector(
		ctx,
		cmdData.RestApiClient,
		uxHelpers.WithOrgPickOnlyOneItem(true),
	)
	if err != nil {
		return entity.Project{}, err
	}

	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected org", selectedOrg.Name.String()))

	label := styles.NewStringBuilder()
	label.WriteString("Type ")
	label.WriteStyledString(
		styles.SelectStyle().
			Bold(true),
		"project",
	)
	label.WriteString(" name")

	name, err := uxBlock.Run(
		input.NewRoot(
			ctx,
			input.WithLabel(label.String()),
			input.WithHelpPlaceholder(),
			input.WithPlaceholderStyle(styles.HelpStyle()),
			input.WithoutPrompt(),
		),
		input.GetValueFunc,
	)
	if err != nil {
		return entity.Project{}, err
	}

	project, err := repository.PostProject(ctx, cmdData.RestApiClient, repository.ProjectPost{
		ClientId: selectedOrg.ID,
		Name:     types.NewString(name),
		Mode:     enum.ProjectModeEnumLight,
	})
	if err != nil {
		return entity.Project{}, err
	}
	cmdData.UxBlocks.PrintSuccessText("Project created")

	return project, nil
}

func ProjectScopeReset(cmdData *LoggedUserCmdData) error {
	_, err := cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.ScopeProjectId = uuid.ProjectIdNull{}
		return data
	})
	if err != nil {
		return err
	}

	return nil
}
