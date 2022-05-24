package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/cliAction/importProjectService"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"
)

func serviceImportCmd() *cobra.Command {
	cmdImport := &cobra.Command{
		Use:          "import [projectNameOrId] [path to import.yml]",
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
			).Import(ctx, importProjectService.RunConfig{
				WorkingDir:      constants.WorkingDir,
				ProjectNameOrId: args[0],
				ImportYamlPath:  args[1],
				ParentCmd:       constants.Service,
			})
		},
	}

	params.RegisterString(cmdImport, "importYamlPath", "", i18n.ImportYamlLocation)
	params.RegisterString(cmdImport, "clientId", "", i18n.ClientId)

	return cmdImport
}
