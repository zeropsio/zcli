package cmd

import (
	"context"

	"time"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/grpcApiClientFactory"

	"github.com/zerops-io/zcli/src/cliAction/startStopDeleteProject"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func deleteProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete project [projectName] --confirm",
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

			region, err := createRegionRetriever()
			if err != nil {
				return err
			}

			reg, err := region.RetrieveFromFile()
			if err != nil {
				return err
			}

			apiClientFactory := grpcApiClientFactory.New(grpcApiClientFactory.Config{
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

			client := httpClient.New(httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			zip := zipClient.New(zipClient.Config{})

			return startStopDeleteProject.New(
				startStopDeleteProject.Config{},
				client,
				zip,
				apiGrpcClient,
			).Run(ctx, startStopDeleteProject.RunConfig{
				ProjectName: args[1],
				Confirm:     params.GetBool(cmd, "confirm"),
			}, constants.Delete)
		},
	}

	params.RegisterBool(cmd, "confirm", false, i18n.ConfirmDeleteProject)

	return cmd
}
