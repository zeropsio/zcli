package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/grpcApiClientFactory"

	"github.com/zerops-io/zcli/src/cliAction/importProject"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func importCmd() *cobra.Command {
	cmd := &cobra.Command{
		// TODO ask how to define voluntary client id var
		Use:          "import project [pathToImportYaml]  --clientId=<string>",
		Short:        i18n.CmdImportDesc,
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

			return importProject.New(
				importProject.Config{},
				client,
				zip,
				apiGrpcClient,
			).ImportProjct(ctx, importProject.RunConfig{
				ZipFilePath:    params.GetString(cmd, "zipFilePath"),
				WorkingDir:     params.GetString(cmd, "workingDir"),
				VersionName:    params.GetString(cmd, "versionName"),
				ImportYamlPath: &args[1],
				ClientId:       params.GetString(cmd, "clientId"),
			})
		},
	}

	params.RegisterString(cmd, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmd, "zipFilePath", "", i18n.BuildZipFilePath)
	params.RegisterString(cmd, "versionName", "", i18n.BuildVersionName)
	params.RegisterString(cmd, "clientId", "", i18n.ClientId)
	params.RegisterString(cmd, "importYamlPath", "", i18n.ImportYamlLocation)

	return cmd
}
