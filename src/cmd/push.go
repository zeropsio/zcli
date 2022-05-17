package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/cliAction/buildDeploy"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"

	"github.com/spf13/cobra"
)

func pushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "push projectName serviceName",
		Short:        i18n.CmdPushDesc,
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
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
			).Push(ctx, buildDeploy.RunConfig{
				ZipFilePath:      params.GetString(cmd, "zipFilePath"),
				WorkingDir:       params.GetString(cmd, "workingDir"),
				VersionName:      params.GetString(cmd, "versionName"),
				SourceName:       params.GetString(cmd, "source"),
				ProjectName:      args[0],
				ServiceStackName: args[1],
			})
		},
	}

	params.RegisterString(cmd, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmd, "zipFilePath", "", i18n.BuildZipFilePath)
	params.RegisterString(cmd, "versionName", "", i18n.BuildVersionName)
	params.RegisterString(cmd, "source", "", i18n.SourceName)

	return cmd
}
