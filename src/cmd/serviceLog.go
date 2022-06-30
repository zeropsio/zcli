package cmd

import (
	"context"
	"github.com/zerops-io/zcli/src/cliAction/serviceLogs"
	"time"

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
		Long:         i18n.CmdServiceLogFull + i18n.ServiceLogAdditional,
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}
			token := getToken(storage)

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

	params.RegisterUInt32(cmd, "limit", 100, i18n.LogLimit)
	params.RegisterString(cmd, "minimumSeverity", "1", i18n.LogMinSeverity)
	params.RegisterString(cmd, "messageType", "APPLICATION", i18n.LogMsgType)
	params.RegisterString(cmd, "format", "FULL", i18n.LogFormat)
	params.RegisterString(cmd, "formatTemplate", "", i18n.LogFormatTemplate)
	params.RegisterBool(cmd, "follow", false, i18n.LogFollow)

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
