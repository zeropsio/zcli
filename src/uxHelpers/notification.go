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

func PrintNotificationList(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	out io.Writer,
	orgId uuid.ClientId,
	projectId uuid.ProjectId,
	limit int,
	offset int,
) error {
	notifications, err := repository.GetUserNotificationsByProject(
		ctx,
		restApiClient,
		orgId,
		projectId,
		limit,
		offset,
	)
	if err != nil {
		return err
	}

	if len(notifications) == 0 {
		_, err = fmt.Fprintln(out, i18n.T(i18n.NotificationListEmpty))
		return err
	}

	header, body := createNotificationTableRows(notifications)

	t := table.Render(body, table.WithHeader(header))

	_, err = fmt.Fprintln(out, t)
	return err
}

func createNotificationTableRows(notifications []entity.UserNotification) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("id", "action", "type", "services", "created by", "created", "ack")

	body := table.NewBody()
	for _, notification := range notifications {
		serviceNames := strings.Join(notification.ServiceNames, ", ")
		if serviceNames == "" {
			serviceNames = "-"
		}

		createdBy := notification.CreatedByUser
		if createdBy == "" {
			createdBy = "-"
		}

		ackStatus := "no"
		if notification.Acknowledged.Native() {
			ackStatus = "yes"
		}

		body.AddStringsRow(
			string(notification.Id),
			notification.ActionName.String(),
			notification.Type.String(),
			serviceNames,
			createdBy,
			notification.ActionCreated.Native().Format(styles.DateTimeFormat),
			ackStatus,
		)
	}

	return header, body
}
