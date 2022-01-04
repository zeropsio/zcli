package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/grpcApiClientFactory"

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
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}

			apiClientFactory := grpcApiClientFactory.New(grpcApiClientFactory.Config{
				CaCertificateUrl: params.GetPersistentString(constants.PersistentParamCaCertificateUrl),
			})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				params.GetPersistentString(constants.PersistentParamGrpcApiAddress),
				getToken(storage),
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			httpClient := httpClient.New(httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			zipClient := zipClient.New(zipClient.Config{})

			return buildDeploy.New(
				buildDeploy.Config{},
				httpClient,
				zipClient,
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
