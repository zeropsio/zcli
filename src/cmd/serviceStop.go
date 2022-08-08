package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/cliAction/startStopDelete"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
)

func serviceStopCmd() *cobra.Command {
	cmdStop := &cobra.Command{
		Use:          "stop projectNameOrId serviceName [flags]",
		Short:        i18n.CmdServiceStop,
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

			apiClientFactory := zBusinessZeropsApiProtocol.New(zBusinessZeropsApiProtocol.Config{
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

			handler := startStopDelete.New(
				startStopDelete.Config{},
				client,
				apiGrpcClient,
				sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress},
			)

			cmdData := startStopDelete.CmdType{
				Start:   i18n.ServiceStop,
				Finish:  i18n.ServiceStopped,
				Execute: handler.ServiceStop,
			}

			return handler.Run(ctx, startStopDelete.RunConfig{
				ProjectNameOrId: args[0],
				ServiceName:     args[1],
				ParentCmd:       constants.Service,
				Confirm:         true,
				CmdData:         cmdData,
			})
		},
	}
	cmdStop.Flags().BoolP("help", "h", false, helpText(i18n.ServiceStopHelp))
	return cmdStop
}
