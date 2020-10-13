package cmd

import (
	"context"

	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/grpcDaemonClientFactory"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/cliAction/startVpn"

	"github.com/spf13/cobra"
)

func vpnStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start projectName",
		Short:        i18n.CmdVpnStart,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
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

			apiGrpcClient, apiCloseFunc, err := createApiGrpcClient(ctx, tlsConfig)
			if err != nil {
				return err
			}
			defer apiCloseFunc()

			installer, err := daemonInstaller.New(daemonInstaller.Config{})
			if err != nil {
				return err
			}

			return startVpn.New(
				startVpn.Config{
					GrpcApiAddress: params.GetString("grpcApiAddress"),
					VpnAddress:     params.GetString("vpnApiAddress"),
				},
				apiGrpcClient,
				grpcDaemonClientFactory.New(),
				installer,
			).Run(ctx, startVpn.RunConfig{
				ProjectName: args[0],
				Token:       token,
				Mtu:         params.GetUint32("mtu"),
			})
		},
	}

	params.RegisterUInt32(cmd, "mtu", 1420, "vpn interface MTU")

	return cmd
}
