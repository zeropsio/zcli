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
)

func serviceImportCmd() *cobra.Command {
	cmdImport := &cobra.Command{
		Use:          "import projectNameOrId pathToImportFile [flags]",
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

			return importProjectService.New(importProjectService.Config{}, client, apiGrpcClient).Import(ctx, importProjectService.RunConfig{
				WorkingDir:      constants.WorkingDir,
				ProjectNameOrId: args[0],
				ImportYamlPath:  args[1],
				ParentCmd:       constants.Service,
			})
		},
	}

	return cmdImport
}
