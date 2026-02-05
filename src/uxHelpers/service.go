package uxHelpers

import (
	"context"
	"encoding/json"
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

type serviceSelectorConfig struct {
	createNew bool
}
type ServiceSelectorOption = gn.Option[serviceSelectorConfig]

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
	cfg := gn.ApplyOptions(opts...)

	services, err := repository.GetNonSystemServicesByProject(ctx, restApiClient, project)
	if err != nil {
		return empty, err
	}

	if len(services) == 0 && !cfg.createNew {
		return empty, errors.New(i18n.T(i18n.ServiceSelectorListEmpty))
	}

	header, body := createServiceTableRows(services, cfg.createNew)

	selected, err := uxBlock.Run(
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

type PrintServiceListConfig struct {
	Format string
}

type serviceListJsonOutput struct {
	Services  []serviceJsonItem `json:"services"`
	Processes []processJsonItem `json:"processes"`
}

type serviceJsonItem struct {
	Id                string  `json:"id"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Status            string  `json:"status"`
	AppVersionId      *string `json:"appVersionId"`
	AppVersionCreated *string `json:"appVersionCreated"`
}

type processJsonItem struct {
	Id        string   `json:"id"`
	Action    string   `json:"action"`
	Status    string   `json:"status"`
	Services  []string `json:"services"`
	CreatedBy string   `json:"createdBy"`
	Created   string   `json:"created"`
}

func PrintServiceList(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	out io.Writer,
	project entity.Project,
	config PrintServiceListConfig,
) error {
	services, err := repository.GetNonSystemServicesByProject(ctx, restApiClient, project)
	if err != nil {
		return err
	}

	// Fetch running/pending processes for the project
	processes, err := repository.GetRunningAndPendingProcessesByProject(
		ctx,
		restApiClient,
		project.OrgId,
		project.Id,
	)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		return printServiceListJson(out, services, processes)
	}

	return printServiceListTable(out, services, processes)
}

func printServiceListJson(out io.Writer, services []entity.Service, processes []entity.Process) error {
	output := serviceListJsonOutput{
		Services:  make([]serviceJsonItem, 0, len(services)),
		Processes: make([]processJsonItem, 0, len(processes)),
	}

	for _, svc := range services {
		item := serviceJsonItem{
			Id:     string(svc.Id),
			Name:   svc.Name.String(),
			Type:   string(svc.ServiceTypeId),
			Status: svc.Status.String(),
		}

		if id, ok := svc.ActiveAppVersionId.Get(); ok {
			idStr := string(id)
			item.AppVersionId = &idStr
		}

		if created, ok := svc.ActiveAppVersionCreated.Get(); ok {
			createdStr := created.Native().Format(styles.DateTimeFormat)
			item.AppVersionCreated = &createdStr
		}

		output.Services = append(output.Services, item)
	}

	for _, process := range processes {
		createdBy := process.CreatedByUser
		if process.CreatedBySystem.Native() {
			createdBy = "system"
		}

		output.Processes = append(output.Processes, processJsonItem{
			Id:        string(process.Id),
			Action:    process.ActionName.String(),
			Status:    process.Status.String(),
			Services:  process.ServiceNames,
			CreatedBy: createdBy,
			Created:   process.Created.Native().Format(styles.DateTimeFormat),
		})
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func printServiceListTable(out io.Writer, services []entity.Service, processes []entity.Process) error {
	header, body := createServiceTableRows(services, false)

	t := table.Render(body, table.WithHeader(header))

	_, err := fmt.Fprintln(out, t)
	if err != nil {
		return err
	}

	// Only show processes section if there are any
	if len(processes) > 0 {
		_, err = fmt.Fprintln(out)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, i18n.T(i18n.ServiceListProcessesHeader))
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out)
		if err != nil {
			return err
		}

		processHeader, processBody := createProcessTableRows(processes)
		pt := table.Render(processBody, table.WithHeader(processHeader))
		_, err = fmt.Fprintln(out, pt)
		if err != nil {
			return err
		}
	}

	return nil
}

func createServiceTableRows(services []entity.Service, createNewService bool) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("id", "name", "type", "status", "app version id", "app version created")

	body := table.NewBody()
	for _, svc := range services {
		appVersionId := "-"
		if id, ok := svc.ActiveAppVersionId.Get(); ok {
			appVersionId = string(id)
		}

		appVersionCreated := "-"
		if created, ok := svc.ActiveAppVersionCreated.Get(); ok {
			appVersionCreated = created.Native().Format(styles.DateTimeFormat)
		}

		body.AddStringsRow(
			string(svc.Id),
			svc.Name.String(),
			string(svc.ServiceTypeId),
			svc.Status.String(),
			appVersionId,
			appVersionCreated,
		)
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
