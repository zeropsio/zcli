package cmd

import (
	"context"

	"github.com/zerops-io/zcli/src/command/startVpn"

	"github.com/spf13/cobra"
)

func startVpnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "startVpn projectName",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel, logger)

			sudoers := createSudoers()

			storage, err := createStorage()
			if err != nil {
				return err
			}

			certReader, err := createCertReader()
			if err != nil {
				return err
			}

			tlsConfig, err := createTlsConfig(certReader)
			if err != nil {
				return err
			}

			apiGrpcClient, apiCloseFunc, err := createApiGrpcClient(ctx, tlsConfig)
			if err != nil {
				return err
			}
			defer apiCloseFunc()

			return startVpn.New(
				startVpn.Config{
					VpnAddress: params.GetString("vpnApiAddress"),
					UserId:     certReader.UserId,
				},
				logger,
				apiGrpcClient,
				sudoers,
				storage,
			).Run(ctx, startVpn.RunConfig{
				ProjectName: args[0],
			})
		},
	}

	return cmd
}
