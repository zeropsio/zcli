package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/startStopDelete"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func projectDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [projectName] --confirm",
		Short:        i18n.CmdProjectDelete,
		Args:         cobra.MinimumNArgs(1),
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

			return startStopDelete.New(
				startStopDelete.Config{},
				client,
				zip,
				apiGrpcClient,
			).Run(ctx, startStopDelete.RunConfig{
				ProjectName: args[0],
				Confirm:     params.GetBool(cmd, "confirm"),
				ParentCmd:   constants.Project,
				ChildCmd:    constants.Delete,
			})
		},
	}
	params.RegisterBool(cmd, "confirm", false, i18n.ConfirmDeleteProject)

	return cmd
}
