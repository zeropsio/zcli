package cmdBuilder

import (
	"context"

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
)

type ProjectOption gn.Option[projectScope]

// WithCreateNewProject allows 'create new project' option in selector
func WithCreateNewProject() ProjectOption {
	return func(s *projectScope) {
		s.createNew = true
	}
}

// WithSkipSelectProject allows skip project select
func WithSkipSelectProject() ProjectOption {
	return func(s *projectScope) {
		s.skipSelect = true
	}
}

type projectScope struct {
	createNew  bool
	skipSelect bool
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

	cmdData.ProjectSelector = p.SelectProject
	if p.skipSelect {
		return nil
	}

	// service scope is set - use project from it
	if service, filled := cmdData.Service.Get(); filled {
		project, err := repository.GetProjectById(ctx, cmdData.RestApiClient, service.ProjectId)
		if err == nil {
			cmdData.Project = optional.New(project)
			return nil
		}
		cmdData.Project = optional.New(project)
	}

	// project scope is set (repo profile only)
	if projectId, ok := cmdData.Profile.ScopeProject(); ok {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
		if err != nil {
			if errorsx.Is(err, errorsx.ErrorCode(errorsx.CodeProjectNotFound)) {
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
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, entity.ProjectId(projectId))
		if err != nil {
			return errorsx.Convert(
				err,
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ errorsx.ApiError, metaItemTyped map[string]interface{}) string {
							return i18n.T(i18n.ErrorInvalidProjectId, projectId, metaItemTyped["message"])
						},
					),
				),
			)
		}
		cmdData.Project = optional.New(project)
	}

	if projectId, exists := cmdData.Args[ProjectArgName]; exists && !cmdData.Project.Filled() {
		project, err = repository.GetProjectById(ctx, cmdData.RestApiClient, entity.ProjectId(projectId[0]))
		if err != nil {
			return errorsx.Convert(
				err,
				errorsx.InvalidUserInput(
					"id",
					errorsx.InvalidUserInputErrorMessage(
						func(_ errorsx.ApiError, metaItemTyped map[string]interface{}) string {
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
		project, selected, err := p.SelectProject(ctx, cmdData)
		if err != nil {
			return err
		}
		if selected {
			cmdData.Project = optional.New(project)
		}
	}

	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.SelectedProject), project.Name))

	return nil
}

func (p *projectScope) SelectProject(ctx context.Context, cmdData *LoggedUserCmdData, opts ...uxHelpers.ProjectSelectorOption) (entity.Project, bool, error) {
	selectedOptions := []uxHelpers.ProjectSelectorOption{
		uxHelpers.WithCreateNewProject(p.createNew),
	}
	// A configured client scope restricts the selector to that org's projects.
	if orgId, ok := cmdData.Profile.ScopeClient(); ok {
		org, err := repository.GetOrgById(ctx, cmdData.RestApiClient, orgId)
		if err != nil {
			return entity.Project{}, false, err
		}
		selectedOptions = append(selectedOptions, uxHelpers.WithProjectOrgScope(org))
	}
	selectedOptions = append(selectedOptions, opts...)
	selectedProject, err := uxHelpers.PrintProjectSelector(
		ctx,
		cmdData.RestApiClient,
		selectedOptions...,
	)
	if err != nil {
		return entity.Project{}, false, err
	}

	if selectedProject.Filled() {
		return selectedProject.Some(), true, nil
	} else if terminal.IsTerminal() {
		project, err := createNewProject(ctx, cmdData)
		if err != nil {
			return entity.Project{}, false, err
		}
		return project, true, nil
	}
	return entity.Project{}, false, nil
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

	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected org", selectedOrg.Name))

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

	location, err := uxHelpers.PrintLocationSelector(ctx, cmdData.RestApiClient)
	if err != nil {
		return entity.Project{}, err
	}
	cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected location", location.Name))

	project, err := repository.PostProject(ctx, cmdData.RestApiClient, entity.PostProject{
		OrgId:    selectedOrg.Id,
		Name:     name,
		Mode:     entity.ProjectModeLight,
		Location: optional.New(location.Id),
	})
	if err != nil {
		return entity.Project{}, err
	}
	cmdData.UxBlocks.PrintSuccessText("Project created")

	return project, nil
}

func ProjectScopeReset(cmdData *LoggedUserCmdData) error {
	if !cmdData.Profile.HasRepo() {
		return nil
	}
	return cmdData.Profile.SetScopeProject(optional.Empty[entity.ProjectId]())
}

// ServiceScopeReset clears the persisted service scope on the active repo profile. No-op when
// no repo profile is in place (matches ProjectScopeReset's behaviour - reset commands stay
// safe to invoke even outside a repo).
func ServiceScopeReset(cmdData *LoggedUserCmdData) error {
	if !cmdData.Profile.HasRepo() {
		return nil
	}
	return cmdData.Profile.SetScopeService(optional.Empty[entity.ServiceId]())
}
