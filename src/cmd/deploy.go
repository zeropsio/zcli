package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/cliAction/buildDeploy"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "deploy [projectNameOrId] [serviceName] [space separated files or directories] --versionName=<string>",
		Short:        i18n.CmdDeployDesc,
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
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

			return buildDeploy.New(
				buildDeploy.Config{},
				client,
				zip,
				apiGrpcClient,
			).Deploy(ctx, buildDeploy.RunConfig{
				ZipFilePath:      params.GetString(cmd, "zipFilePath"),
				WorkingDir:       constants.WorkingDir,
				VersionName:      params.GetString(cmd, "versionName"),
				ZeropsYamlPath:   params.GetStringP(cmd, "zeropsYamlPath"),
				ProjectNameOrId:  args[0],
				ServiceStackName: args[1],
				PathsForPacking:  args[2:],
			})
		},
	}

	// TODO review flags
	params.RegisterString(cmd, "zipFilePath", "", i18n.BuildZipFilePath)
	params.RegisterString(cmd, "versionName", "", i18n.BuildVersionName)
	params.RegisterString(cmd, "zeropsYamlPath", "./", i18n.ZeropsYamlLocation)

	return cmd
}
