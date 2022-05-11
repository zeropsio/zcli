package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/importProjectService"
	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/constants"

	"github.com/zerops-io/zcli/src/cliAction/startStopDeleteProject"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func serviceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "service", Short: i18n.CmdService}

	cmdStart := &cobra.Command{
		Use:          "start [projectName] [serviceName]",
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

	cmdStop := &cobra.Command{
		Use:          "stop [projectName] [serviceName]",
		Short:        i18n.CmdServiceStop,
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
			}, constants.Stop)
		},
	}
	cmdDelete := &cobra.Command{
		Use:          "delete [projectName] [serviceName] --confirm",
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

	cmdImport := &cobra.Command{
		Use:          "import [projectName] [path to import.yml]",
		Short:        i18n.CmdServiceImport,
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

			return importProjectService.New(
				importProjectService.Config{},
				client,
				zip,
				apiGrpcClient,
			).ImportService(ctx, importProjectService.RunConfig{
				WorkingDir:     params.GetString(cmd, "workingDir"),
				ProjectName:    args[0],
				ImportYamlPath: &args[1],
			})
		},
	}

	params.RegisterBool(cmdDelete, "confirm", false, i18n.ConfirmDeleteService)

	params.RegisterString(cmdImport, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmdImport, "importYamlPath", "", i18n.ImportYamlLocation)
	params.RegisterString(cmdImport, "clientId", "", i18n.ClientId)

	cmd.AddCommand(cmdStart, cmdStop, cmdDelete, cmdImport)
	return cmd
}
