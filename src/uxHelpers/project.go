package uxHelpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

func PrintProjectSelector(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler,
) (*entity.Project, error) {
	projects, err := repository.GetAllProjects(ctx, restApiClient)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.ProjectSelectorListEmpty)))
		return nil, nil
	}

	header, tableBody := createProjectTableRows(projects)

	projectIndex, err := uxBlocks.Select(
		ctx,
		tableBody,
		uxBlock.SelectLabel(i18n.T(i18n.ProjectSelectorPrompt)),
		uxBlock.SelectTableHeader(header),
	)
	if err != nil {
		return nil, err
	}

	if len(projectIndex) == 0 {
		return nil, errors.New(i18n.T(i18n.ProjectSelectorOutOfRangeError))
	}

	if projectIndex[0] > len(projects)-1 {
		return nil, errors.New(i18n.T(i18n.ProjectSelectorOutOfRangeError))
	}

	return &projects[projectIndex[0]], nil
}

func PrintProjectList(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler) error {
	projects, err := repository.GetAllProjects(ctx, restApiClient)
	if err != nil {
		return err
	}

	header, rows := createProjectTableRows(projects)

	uxBlocks.Table(rows, uxBlock.WithTableHeader(header))

	return nil
}

func createProjectTableRows(projects []entity.Project) (*uxBlock.TableRow, *uxBlock.TableBody) {
	// TODO - janhajek translation
	header := (&uxBlock.TableRow{}).AddStringCells("ID", "Name", "Description", "Org ID", "Status")

	tableBody := &uxBlock.TableBody{}
	for _, project := range projects {
		tableBody.AddStringsRow(
			string(project.ID),
			project.Name.String(),
			project.Description.Native(),
			project.ClientId.Native(),
			project.Status.String(),
		)
	}

	return header, tableBody
}
