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
		Short(i18n.T(i18n.CmdServiceLog)).
		Long(i18n.T(i18n.CmdServiceLogLong)+i18n.T(i18n.ServiceLogAdditional)).
		ScopeLevel(cmdBuilder.Service).
		StringFlag("x-amz-acl", "", i18n.T(i18n.BucketGenericXAmzAcl)).
		IntFlag("limit", 100, i18n.T(i18n.LogLimitFlag)).
		StringFlag("minimumSeverity", "", i18n.T(i18n.LogMinSeverityFlag)).
		StringFlag("messageType", "APPLICATION", i18n.T(i18n.LogMsgTypeFlag)).
		StringFlag("format", "FULL", i18n.T(i18n.LogFormatFlag)).
		StringFlag("formatTemplate", "", i18n.T(i18n.LogFormatTemplateFlag)).
		BoolFlag("follow", false, i18n.T(i18n.LogFollowFlag)).
		BoolFlag("showBuildLogs", false, i18n.T(i18n.LogShowBuildFlag)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			handler := serviceLogs.New(
				serviceLogs.Config{},
				cmdData.RestApiClient,
			)

			serviceId := cmdData.Service.ID
			if cmdData.Params.GetBool("showBuildLogs") {
				appVersions, err := repository.GetLatestAppVersionByService(ctx, cmdData.RestApiClient, *cmdData.Service)
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
				Project:        *cmdData.Project,
				ServiceId:      serviceId,
				Limit:          uint32(cmdData.Params.GetInt("limit")),
				MinSeverity:    cmdData.Params.GetString("minimumSeverity"),
				MsgType:        cmdData.Params.GetString("messageType"),
				Format:         cmdData.Params.GetString("format"),
				FormatTemplate: cmdData.Params.GetString("formatTemplate"),
				Follow:         cmdData.Params.GetBool("follow"),
				// TODO - janhajek better place?
				Levels: serviceLogs.Levels{
					{"EMERGENCY", "0"},
					{"ALERT", "1"},
					{"CRITICAL", "2"},
					{"ERROR", "3"},
					{"WARNING", "4"},
					{"NOTICE", "5"},
					{"INFORMATIONAL", "6"},
					{"DEBUG", "7"},
				},
			})
		})
}
