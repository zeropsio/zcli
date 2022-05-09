package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/importProjectService"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func importCmd() *cobra.Command {

	cmd := &cobra.Command{Use: "import", Short: "import project or service"}
	cmdProject := &cobra.Command{
		// TODO ask how to define voluntary client id var
		Use:          "project [pathToImportYaml]  --clientId=<string>",
		Short:        i18n.CmdProjectImport,
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

			return importProjectService.New(
				importProjectService.Config{},
				client,
				zip,
				apiGrpcClient,
			).Run(ctx, importProjectService.RunConfig{
				// 				ZipFilePath:    params.GetString(cmd, "zipFilePath"),
				WorkingDir: params.GetString(cmd, "workingDir"),
				// 				VersionName:    params.GetString(cmd, "versionName"),
				ImportYamlPath: &args[1],
				ClientId:       params.GetString(cmd, "clientId"),
			})
		},
	}

	cmdService := &cobra.Command{
		Use:          "service [projectName] [path to import.yml]",
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

	params.RegisterString(cmd, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmd, "importYamlPath", "", i18n.ImportYamlLocation)
	params.RegisterString(cmdProject, "clientId", "", i18n.ClientId)

	cmd.AddCommand(cmdProject, cmdService)
	return cmd
}
