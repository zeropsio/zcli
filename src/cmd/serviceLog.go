package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/serviceLogs"
	"github.com/zeropsio/zerops-go/types/enum"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceLogCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("log").
		Short(i18n.T(i18n.CmdDescServiceLog)).
		Long(i18n.T(i18n.CmdDescServiceLogLong)).
		ScopeLevel(cmdBuilder.ScopeService()).
		IntFlag("limit", 100, i18n.T(i18n.LogLimitFlag)).
		StringFlag("minimum-severity", "", i18n.T(i18n.LogMinSeverityFlag)).
		StringFlag("message-type", "APPLICATION", i18n.T(i18n.LogMsgTypeFlag)).
		StringFlag("format", "FULL", i18n.T(i18n.LogFormatFlag)).
		StringFlag("format-template", "", i18n.T(i18n.LogFormatTemplateFlag)).
		BoolFlag("follow", false, i18n.T(i18n.LogFollowFlag)).
		BoolFlag("show-build-logs", false, i18n.T(i18n.LogShowBuildFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpServiceLog)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			handler := serviceLogs.NewStdout(
				serviceLogs.Config{},
				cmdData.RestApiClient,
			)

			serviceId := service.ID
			if cmdData.Params.GetBool("showBuildLogs") {
				appVersions, err := repository.GetLatestAppVersionByService(ctx, cmdData.RestApiClient, service)
				if err != nil {
					return err
				}
				if len(appVersions) == 0 {
					return errors.New(i18n.T(i18n.LogNoBuildFound))
				}

				app := appVersions[0]
				status := app.Status
				if status == enum.AppVersionStatusEnumUploading || app.Build == nil {
					return errors.New(i18n.T(i18n.LogBuildStatusUploading))
				}

				var filled bool
				serviceId, filled = app.Build.ServiceStackId.Get()
				if !filled {
					return errors.New(i18n.T(i18n.LogNoBuildFound))
				}
			}

			return handler.Run(ctx, serviceLogs.RunConfig{
				Project:        project,
				ServiceId:      serviceId,
				Limit:          cmdData.Params.GetInt("limit"),
				MinSeverity:    cmdData.Params.GetString("minimum-severity"),
				MsgType:        cmdData.Params.GetString("message-type"),
				Format:         cmdData.Params.GetString("format"),
				FormatTemplate: cmdData.Params.GetString("format-template"),
				Follow:         cmdData.Params.GetBool("follow"),
				// TODO - janhajek better place?
				Levels: serviceLogs.DefaultLevels,
			})
		})
}
