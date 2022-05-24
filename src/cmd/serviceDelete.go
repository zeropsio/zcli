package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/constants"

	"github.com/zerops-io/zcli/src/cliAction/startStopDelete"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func serviceDeleteCmd() *cobra.Command {
	cmdDelete := &cobra.Command{
		Use:          "delete [projectNameOrId] [serviceName] --confirm",
		Short:        i18n.CmdServiceDelete,
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			storage, err := createCliStorage()
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
				getToken(storage),
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			client := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			zip := zipClient.New(zipClient.Config{})

			handler := startStopDelete.New(
				startStopDelete.Config{},
				client,
				zip,
				apiGrpcClient,
			)

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

	params.RegisterBool(cmdDelete, "confirm", false, i18n.ConfirmDeleteService)
	return cmdDelete
}
