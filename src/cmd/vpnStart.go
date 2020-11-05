package cmd

import (
	"context"

	"github.com/zerops-io/zcli/src/grpcApiClientFactory"

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
			apiClientFactory := grpcApiClientFactory.New(grpcApiClientFactory.Config{
				CaCertificate: params.GetPersistentBytes("caCertificate"),
			})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				params.GetPersistentString("grpcApiAddress"),
				token,
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			installer, err := daemonInstaller.New(daemonInstaller.Config{})
			if err != nil {
				return err
			}

			return startVpn.New(
				startVpn.Config{
					GrpcApiAddress: params.GetPersistentString("grpcApiAddress"),
					VpnAddress:     params.GetPersistentString("vpnApiAddress"),
				},
				apiGrpcClient,
				grpcDaemonClientFactory.New(),
				installer,
			).Run(ctx, startVpn.RunConfig{
				ProjectName:   args[0],
				Token:         token,
				Mtu:           params.GetUint32("mtu"),
				CaCertificate: params.GetPersistentBytes("caCertificate"),
			})
		},
	}

	params.RegisterUInt32(cmd, "mtu", 1420, "vpn interface MTU")

	return cmd
}
