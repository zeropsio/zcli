package uxHelpers

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

type projectSelectorConfig struct {
	createNew bool
}
type ProjectSelectorOption = generic.Option[projectSelectorConfig]

func WithCreateNewProject(b bool) ProjectSelectorOption {
	return func(s *projectSelectorConfig) {
		s.createNew = b
	}
}

func PrintProjectSelector(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	opts ...ProjectSelectorOption,
) (optional.Null[entity.Project], error) {
	var empty optional.Null[entity.Project]
	cfg := generic.ApplyOptions(opts...)

	projects, err := repository.GetAllProjects(ctx, restApiClient)
	if err != nil {
		return empty, err
	}

	if len(projects) == 0 && !cfg.createNew {
		return empty, errors.New(i18n.T(i18n.ProjectSelectorListEmpty))
	}

	header, body := createProjectTableRows(projects, cfg.createNew)

	selected, err := uxBlock.RunR(
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
	header := table.NewRowFromStrings("ID", "Name", "Org Name", "Org ID", "Status")

	tableBody := table.NewBody()
	for _, project := range projects {
		tableBody.AddStringsRow(
			string(project.ID),
			project.Name.String(),
			project.OrgName.Native(),
			project.OrgId.Native(),
			project.Status.String(),
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
