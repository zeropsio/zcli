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
		Use:          "deploy projectName serviceName",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(2),
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

			zipClient := zipClient.New(zipClient.Config{})

			return deploy.New(
				deploy.Config{},
				httpClient,
				zipClient,
				logger,
				apiGrpcClient,
			).Run(ctx, deploy.RunConfig{
				SourceDirectoryPath: params.GetString("sourceDirectory"),
				ProjectName:         args[0],
				ServiceStackName:    args[1],
			})
		},
	}

	params.RegisterString(cmd, "sourceDirectory", "./", "directory with source code, it will be zipped and will be send to zerops server for deploy")

	return cmd
}
