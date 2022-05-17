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

func projectImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		// TODO ask how to define voluntary client id var
		Use:          "import [pathToImportYaml]  --clientId=<string>",
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

			return importProjectService.New(
				importProjectService.Config{},
				client,
				zip,
				apiGrpcClient,
			).Run(ctx, importProjectService.RunConfig{
				// 				ZipFilePath:    params.GetString(cmd, "zipFilePath"),
				WorkingDir: params.GetString(cmd, "workingDir"),
				// 				VersionName:    params.GetString(cmd, "versionName"),
				ImportYamlPath: &args[0],
				ClientId:       params.GetString(cmd, "clientId"),
			})
		},
	}

	params.RegisterString(cmd, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmd, "importYamlPath", "", i18n.ImportYamlLocation)
	params.RegisterString(cmd, "clientId", "", i18n.ClientId)

	return cmd
}
