package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/cliAction/startStopDelete"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
)

func serviceDeleteCmd() *cobra.Command {
	cmdDelete := &cobra.Command{
		Use:          "delete projectNameOrId serviceName [flags]",
		Short:        i18n.CmdServiceDelete,
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

			handler := startStopDelete.New(startStopDelete.Config{}, client, apiGrpcClient, sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress})

			cmdData := startStopDelete.CmdType{
				Start:   i18n.ServiceDelete,
				Finish:  i18n.ServiceDeleted,
				Execute: handler.ServiceDelete,
			}

			return handler.Run(ctx, startStopDelete.RunConfig{
				ProjectNameOrId: args[0],
				ServiceName:     args[1],
				Confirm:         params.GetBool(cmd, "confirm"),
				ParentCmd:       constants.Service,
				CmdData:         cmdData,
			})
		},
	}

	params.RegisterBool(cmdDelete, "confirm", false, i18n.ConfirmDelete)
	cmdDelete.Flags().BoolP("help", "h", false, helpText(i18n.ServiceDeleteHelp))

	return cmdDelete
}
