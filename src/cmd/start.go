package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/constants"

	"github.com/zerops-io/zcli/src/cliAction/startStopDeleteProject"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "start", Short: i18n.CmdStart}
	cmdProject := &cobra.Command{
		Use:          "project [projectName]",
		Short:        i18n.CmdProjectStart,
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
				ProjectName: args[0],
			}, constants.Start)
		},
	}
	cmdService := &cobra.Command{
		Use:          "service [projectName] [serviceName]",
		Short:        i18n.CmdServiceStart,
		Args:         cobra.MinimumNArgs(2),
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
				ProjectName: args[0],
				ServiceName: args[1],
			}, constants.Start)
		},
	}

	cmd.AddCommand(cmdProject, cmdService)
	return cmd
}
