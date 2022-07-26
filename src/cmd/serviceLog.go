package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/serviceLogs"

	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"

	"github.com/spf13/cobra"
)

func serviceLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "log projectNameOrId serviceName [flags]",
		Short:        i18n.CmdServiceLog,
		Long:         i18n.CmdServiceLogLong + i18n.ServiceLogAdditional,
		Args:         ExactNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}
			token, err := getToken(storage)
			if err != nil {
				return err
			}

			region, err := createRegionRetriever(ctx)
			if err != nil {
				return err
			}

			reg, err := region.RetrieveFromFile()
			if err != nil {
				return err
			}

			apiClientFactory := business.New(business.Config{
				CaCertificateUrl: reg.CaCertificateUrl,
			})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				reg.GrpcApiAddress,
				token,
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			client := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			handler := serviceLogs.New(serviceLogs.Config{}, client, apiGrpcClient, sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress})

			severityLevel := serviceLogs.Levels{{"EMERGENCY", "0"}, {"ALERT", "1"}, {"CRITICAL", "2"}, {"ERROR", "3"}, {"WARNING", "4"}, {"NOTICE", "5"}, {"INFORMATIONAL", "6"}, {"DEBUG", "7"}}

			return handler.Run(ctx, serviceLogs.RunConfig{
				ProjectNameOrId: args[0],
				ServiceName:     args[1],
				Limit:           params.GetUint32("limit"),
				MinSeverity:     params.GetString(cmd, "minimumSeverity"),
				MsgType:         params.GetString(cmd, "messageType"),
				Format:          params.GetString(cmd, "format"),
				FormatTemplate:  params.GetString(cmd, "formatTemplate"),
				Follow:          params.GetBool(cmd, "follow"),
				Levels:          severityLevel,
			})
		},
	}

	params.RegisterUInt32(cmd, "limit", 100, i18n.LogLimitFlag)
	params.RegisterString(cmd, "minimumSeverity", "", i18n.LogMinSeverityFlag)
	params.RegisterString(cmd, "messageType", "APPLICATION", i18n.LogMsgTypeFlag)
	params.RegisterString(cmd, "format", "FULL", i18n.LogFormatFlag)
	params.RegisterString(cmd, "formatTemplate", "", i18n.LogFormatTemplateFlag)
	params.RegisterBool(cmd, "follow", false, i18n.LogFollowFlag)

	cmd.Flags().BoolP("help", "h", false, helpText(i18n.ServiceLogHelp))

	// TODO remove when websocket is enabled
	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		err := command.Flags().MarkHidden("follow")
		if err != nil {
			return
		}
		command.Parent().HelpFunc()(command, strings)
	})

	return cmd
}
