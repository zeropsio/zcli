package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func projectNotificationsCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("notifications").
		Short(i18n.T(i18n.CmdDescProjectNotifications)).
		Long(i18n.T(i18n.CmdDescProjectNotificationsLong)).
		ScopeLevel(cmdBuilder.ScopeProject()).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		IntFlag("limit", 50, i18n.T(i18n.NotificationLimitFlag)).
		IntFlag("offset", 0, i18n.T(i18n.NotificationOffsetFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpProjectNotifications)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			limit := cmdData.Params.GetInt("limit")
			offset := cmdData.Params.GetInt("offset")

			// Validate limit
			if limit < 1 || limit > 100 {
				return errors.New(i18n.T(i18n.NotificationLimitInvalid))
			}

			// Validate offset
			if offset < 0 {
				return errors.New(i18n.T(i18n.NotificationOffsetInvalid))
			}

			return uxHelpers.PrintNotificationList(
				ctx,
				cmdData.RestApiClient,
				cmdData.Stdout,
				project.OrgId,
				project.Id,
				limit,
				offset,
			)
		})
}
