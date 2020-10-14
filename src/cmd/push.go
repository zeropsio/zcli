package cmd

import (
	"context"
	"time"

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

			token := getToken(storage)

			certReader, err := createCertReader(token)
			if err != nil {
				return err
			}

			tlsConfig, err := createTlsConfig(certReader)
			if err != nil {
				return err
			}

			apiGrpcClient, closeFunc, err := createApiGrpcClient(ctx, tlsConfig)
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
				ProjectName:      args[0],
				ServiceStackName: args[1],
			})
		},
	}

	params.RegisterString(cmd, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmd, "zipFilePath", "", i18n.BuildZipFilePath)
	params.RegisterString(cmd, "versionName", "", i18n.BuildVersionName)

	return cmd
}
