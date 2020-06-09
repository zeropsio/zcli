package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/command/deploy"
	"github.com/zerops-io/zcli/src/service/httpClient"
	"github.com/zerops-io/zcli/src/service/zipClient"

	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "deploy projectName serviceName pathToFileOrDir [pathToFileOrDir]",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel, logger)

			certReader, err := createCertReader()
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

			zipClient := zipClient.New(zipClient.Config{}, logger)

			return deploy.New(
				deploy.Config{},
				httpClient,
				zipClient,
				logger,
				apiGrpcClient,
			).Run(ctx, deploy.RunConfig{
				ZipFilePath:      params.GetString("zipFilePath"),
				WorkingDir:       params.GetString("workingDir"),
				ProjectName:      args[0],
				ServiceStackName: args[1],
				PathsForPacking:  args[2:],
			})
		},
	}

	params.RegisterString(cmd, "workingDir", "./", "working dir, all files path are relative to this directory")
	params.RegisterString(cmd, "zipFilePath", "", "if it's set, save final zip file")

	return cmd
}
