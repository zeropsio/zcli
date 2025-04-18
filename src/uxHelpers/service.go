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

type serviceSelectorConfig struct {
	createNew bool
}
type ServiceSelectorOption = generic.Option[serviceSelectorConfig]

func WithCreateNewService(b bool) ServiceSelectorOption {
	return func(s *serviceSelectorConfig) {
		s.createNew = b
	}
}

func PrintServiceSelector(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	project entity.Project,
	opts ...ServiceSelectorOption,
) (optional.Null[entity.Service], error) {
	var empty optional.Null[entity.Service]
	cfg := generic.ApplyOptions(opts...)

	services, err := repository.GetNonSystemServicesByProject(ctx, restApiClient, project)
	if err != nil {
		return empty, err
	}

	if len(services) == 0 && !cfg.createNew {
		return empty, errors.New(i18n.T(i18n.ServiceSelectorListEmpty))
	}

	header, body := createServiceTableRows(services, cfg.createNew)

	selected, err := uxBlock.RunR(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel(i18n.T(i18n.ServiceSelectorPrompt)),
			selector.WithHeader(header),
			selector.WithSetEnableFiltering(true),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return empty, err
	}

	if selected <= len(services)-1 {
		return optional.New(services[selected]), nil
	}

	return empty, nil
}

func PrintServiceList(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	out io.Writer,
	project entity.Project,
) error {
	services, err := repository.GetNonSystemServicesByProject(ctx, restApiClient, project)
	if err != nil {
		return err
	}

	header, body := createServiceTableRows(services, false)

	t := table.Render(body, table.WithHeader(header))

	_, err = fmt.Fprintln(out, t)
	return err
}

func createServiceTableRows(projects []entity.Service, createNewService bool) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("ID", "Name", "Status")

	body := table.NewBody()
	for _, project := range projects {
		body.AddStringsRow(string(project.ID), project.Name.String(), project.Status.String())
	}
	if createNewService {
		body.AddCellsRow(
			table.NewCell("Create new service").
				SetStyle(
					styles.DefaultStyle().
						Foreground(styles.GreenColor).
						Bold(true),
				).
				SetPretty(true),
		)
	}

	return header, body
}
