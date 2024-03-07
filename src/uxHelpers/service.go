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

func PrintServiceSelector(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler,
	project entity.Project,
) (*entity.Service, error) {
	services, err := repository.GetNonSystemServicesByProject(ctx, restApiClient, project)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.ServiceSelectorListEmpty)))
		return nil, nil
	}

	header, rows := createServiceTableRows(services)

	serviceIndex, err := uxBlocks.Select(
		ctx,
		rows,
		uxBlock.SelectLabel(i18n.T(i18n.ServiceSelectorPrompt)),
		uxBlock.SelectTableHeader(header),
	)
	if err != nil {
		return nil, err
	}

	if len(serviceIndex) == 0 {
		return nil, errors.New(i18n.T(i18n.ServiceSelectorOutOfRangeError))
	}

	if serviceIndex[0] > len(services)-1 {
		return nil, errors.New(i18n.T(i18n.ServiceSelectorOutOfRangeError))
	}

	return &services[serviceIndex[0]], nil
}

func PrintServiceList(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler,
	project entity.Project,
) error {
	services, err := repository.GetNonSystemServicesByProject(ctx, restApiClient, project)
	if err != nil {
		return err
	}

	header, tableBody := createServiceTableRows(services)

	uxBlocks.Table(tableBody, uxBlock.WithTableHeader(header))

	return nil
}

func createServiceTableRows(projects []entity.Service) (*uxBlock.TableRow, *uxBlock.TableBody) {
	// TODO - janhajek translation
	header := (&uxBlock.TableRow{}).AddStringCells("ID", "Name", "Status")

	tableBody := &uxBlock.TableBody{}
	for _, project := range projects {
		tableBody.AddStringsRow(string(project.ID), project.Name.String(), project.Status.String())
	}

	return header, tableBody
}
