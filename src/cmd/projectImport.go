package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/importProjectService"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"

	"github.com/spf13/cobra"
)

func projectImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "import projectNameOrId pathToImportFile [flags]",
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
			token := getToken(storage)

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
				token,
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			client := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			return importProjectService.New(importProjectService.Config{}, client, apiGrpcClient, token).Import(ctx, importProjectService.RunConfig{
				WorkingDir:     constants.WorkingDir,
				ImportYamlPath: args[0],
				ClientId:       params.GetString(cmd, "clientId"),
				ParentCmd:      constants.Project,
			})
		},
	}

	params.RegisterString(cmd, "clientId", "", i18n.ClientId)

	return cmd
}
