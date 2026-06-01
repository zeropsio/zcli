package uxHelpers

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

type projectSelectorConfig struct {
	preselectedId string
	createNew     bool
	orgScope      optional.Null[entity.Org]
}
type ProjectSelectorOption = gn.Option[projectSelectorConfig]

func WithCreateNewProject(b bool) ProjectSelectorOption {
	return func(s *projectSelectorConfig) {
		s.createNew = b
	}
}

// WithProjectOrgScope restricts the selector to a single org's projects. Used when a client
// scope is in effect so only that client's projects are fetched and shown.
func WithProjectOrgScope(org entity.Org) ProjectSelectorOption {
	return func(s *projectSelectorConfig) {
		s.orgScope = optional.New(org)
	}
}

func WithPreselectedProjectId(projectId string) ProjectSelectorOption {
	return func(s *projectSelectorConfig) {
		s.preselectedId = projectId
	}
}

func PrintProjectSelector(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	opts ...ProjectSelectorOption,
) (optional.Null[entity.Project], error) {
	var empty optional.Null[entity.Project]
	cfg := gn.ApplyOptions(opts...)

	projects, err := getSelectorProjects(ctx, restApiClient, cfg.orgScope)
	if err != nil {
		return empty, err
	}

	if cfg.preselectedId != "" {
		var preselectedName []entity.Project
		for _, project := range projects {
			if project.Id.String() == cfg.preselectedId {
				preselectedName = append(preselectedName, project)
			}
			if project.Name == cfg.preselectedId {
				preselectedName = append(preselectedName, project)
			}
		}
		if len(preselectedName) == 1 {
			return optional.New(preselectedName[0]), nil
		}
		projects = preselectedName
	}

	if len(projects) == 0 && !cfg.createNew {
		return empty, errors.New(i18n.T(i18n.ProjectSelectorListEmpty))
	}

	header, body := createProjectTableRows(projects, cfg.createNew)

	selected, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel(i18n.T(i18n.ProjectSelectorPrompt)),
			selector.WithHeader(header),
			selector.WithSetEnableFiltering(true),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return empty, err
	}

	if selected <= len(projects)-1 {
		return optional.New(projects[selected]), nil
	}

	return empty, nil
}

// getSelectorProjects fetches the projects the selector should offer: only the scoped org's
// projects when a client scope is set, otherwise every org's projects.
func getSelectorProjects(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgScope optional.Null[entity.Org],
) ([]entity.Project, error) {
	if org, ok := orgScope.Get(); ok {
		return repository.GetProjectsByOrg(ctx, restApiClient, org)
	}
	return repository.GetAllProjects(ctx, restApiClient)
}

func PrintProjectList(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	out io.Writer,
) error {
	projects, err := repository.GetAllProjects(ctx, restApiClient)
	if err != nil {
		return err
	}

	header, body := createProjectTableRows(projects, false)

	t := table.Render(body, table.WithHeader(header))

	_, err = fmt.Fprintln(out, t)
	return err
}

func createProjectTableRows(projects []entity.Project, createNewProject bool) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("id", "name", "org name", "org id", "status", "mode")

	tableBody := table.NewBody()
	for _, project := range projects {
		tableBody.AddStringsRow(
			project.Id.String(),
			project.Name,
			project.OrgName,
			project.OrgId.String(),
			project.Status.String(),
			project.Mode.String(),
		)
	}
	if createNewProject {
		tableBody.AddCellsRow(
			table.NewCell("Create new project").
				SetStyle(
					styles.DefaultStyle().
						Foreground(styles.GreenColor).
						Bold(true),
				).
				SetPretty(true),
		)
	}

	return header, tableBody
}
