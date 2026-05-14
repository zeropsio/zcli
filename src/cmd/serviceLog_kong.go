package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/serviceLogs"
	"github.com/zeropsio/zerops-go/types/enum"
)

// ServiceLogCmd streams logs for the scoped service. Reading the typed
// struct field for ShowBuildLogs (instead of the buggy cmdData.Params
// .GetBool("showBuildLogs") in cobra) is the only behavioural difference.
type ServiceLogCmd struct {
	kongCmd.ServiceScope

	Limit           int    `name:"limit" default:"100" help:"Maximum number of log lines to retrieve."`
	MinimumSeverity string `name:"minimum-severity" help:"Minimum severity level."`
	MessageType     string `name:"message-type" default:"APPLICATION" help:"Type of messages to retrieve."`
	Format          string `name:"format" default:"FULL" help:"Output format."`
	FormatTemplate  string `name:"format-template" help:"Custom Go template for output formatting."`
	Follow          bool   `name:"follow" help:"Stream logs continuously."`
	ShowBuildLogs   bool   `name:"show-build-logs" help:"Show logs from the latest build instead of runtime logs."`
}

func (c *ServiceLogCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	service, err := auth.RequireService(ctx, c.ServiceScope)
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	handler := serviceLogs.NewStdout(serviceLogs.Config{}, auth.RestApiClient)

	serviceID := service.Id
	if c.ShowBuildLogs {
		appVersions, err := repository.GetLatestAppVersionByService(ctx, auth.RestApiClient, service)
		if err != nil {
			return err
		}
		if len(appVersions) == 0 {
			return errors.New(i18n.T(i18n.LogNoBuildFound))
		}
		app := appVersions[0]
		if app.Status == enum.AppVersionStatusEnumUploading || app.Build == nil {
			return errors.New(i18n.T(i18n.LogBuildStatusUploading))
		}
		var filled bool
		serviceID, filled = app.Build.ServiceStackId.Get()
		if !filled {
			return errors.New(i18n.T(i18n.LogNoBuildFound))
		}
	}

	return handler.Run(ctx, serviceLogs.RunConfig{
		Project:        project,
		ServiceId:      serviceID,
		Limit:          c.Limit,
		MinSeverity:    c.MinimumSeverity,
		MsgType:        c.MessageType,
		Format:         c.Format,
		FormatTemplate: c.FormatTemplate,
		Follow:         c.Follow,
		Levels:         serviceLogs.DefaultLevels,
	})
}
