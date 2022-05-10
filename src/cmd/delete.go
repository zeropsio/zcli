package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/startStopDeleteProject"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func deleteCmd() *cobra.Command {
	//todo add to dict
	cmd := &cobra.Command{Use: "delete", Short: i18n.CmdDelete}
	cmdProject := &cobra.Command{
		Use:          "project [projectName] --confirm",
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
				Confirm:     params.GetBool(cmd, "confirm"),
			}, constants.Delete)
		},
	}
	cmdService := &cobra.Command{
		Use:          "service [projectName] [serviceName] --confirm",
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
				Confirm:     params.GetBool(cmd, "confirm"),
			}, constants.Delete)
		},
	}

	params.RegisterBool(cmdProject, "confirm", false, i18n.ConfirmDeleteProject)
	params.RegisterBool(cmdService, "confirm", false, i18n.ConfirmDeleteService)

	cmd.AddCommand(cmdProject, cmdService)
	return cmd
}
