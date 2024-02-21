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

func PrintOrgSelector(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler,
) (*entity.Org, error) {
	orgs, err := repository.GetAllOrgs(ctx, restApiClient)
	if err != nil {
		return nil, err
	}

	if len(orgs) == 0 {
		uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.OrgSelectorListEmpty)))
		return nil, nil
	}

	header, tableBody := createOrgTableRows(orgs)

	orgIndex, err := uxBlocks.Select(
		ctx,
		tableBody,
		uxBlock.SelectLabel(i18n.T(i18n.OrgSelectorPrompt)),
		uxBlock.SelectTableHeader(header),
	)
	if err != nil {
		return nil, err
	}

	if len(orgIndex) == 0 {
		return nil, errors.New(i18n.T(i18n.OrgSelectorOutOfRangeError))
	}

	if orgIndex[0] > len(orgs)-1 {
		return nil, errors.New(i18n.T(i18n.OrgSelectorOutOfRangeError))
	}

	return &orgs[orgIndex[0]], nil
}

func createOrgTableRows(projects []entity.Org) (*uxBlock.TableRow, *uxBlock.TableBody) {
	// TODO - janhajek translation
	header := (&uxBlock.TableRow{}).AddStringCells("ID", "Name", "Role")

	tableBody := &uxBlock.TableBody{}
	for _, project := range projects {
		tableBody.AddStringsRow(
			string(project.ID),
			project.Name.String(),
			project.Role.Native(),
		)
	}

	return header, tableBody
}
