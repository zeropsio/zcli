package uxHelpers

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func PrintProcessList(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	out io.Writer,
	orgId uuid.ClientId,
	projectId uuid.ProjectId,
) error {
	processes, err := repository.GetRunningAndPendingProcessesByProject(
		ctx,
		restApiClient,
		orgId,
		projectId,
	)
	if err != nil {
		return err
	}

	if len(processes) == 0 {
		_, err = fmt.Fprintln(out, i18n.T(i18n.ProcessListEmpty))
		return err
	}

	header, body := createProcessTableRows(processes)

	t := table.Render(body, table.WithHeader(header))

	_, err = fmt.Fprintln(out, t)
	return err
}

func createProcessTableRows(processes []entity.Process) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("id", "action", "status", "services", "created by", "created")

	body := table.NewBody()
	for _, process := range processes {
		serviceNames := strings.Join(process.ServiceNames, ", ")
		if serviceNames == "" {
			serviceNames = "-"
		}

		createdBy := process.CreatedByUser
		if process.CreatedBySystem.Native() {
			createdBy = "system"
		}
		if createdBy == "" {
			createdBy = "-"
		}

		body.AddStringsRow(
			string(process.Id),
			process.ActionName.String(),
			process.Status.String(),
			serviceNames,
			createdBy,
			process.Created.Native().Format(styles.DateTimeFormat),
		)
	}

	return header, body
}
