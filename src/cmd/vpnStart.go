package cmd

import (
	"context"

	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"

	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/cliAction/startVpn"

	"github.com/spf13/cobra"
)

func vpnStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start projectNameOrId",
		Short:        i18n.CmdVpnStart,
		SilenceUsage: true,
		Args:         CustomMessageArgs(cobra.ExactArgs(1), i18n.VpnStartExpectedProjectName),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
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

			caCertUrl := reg.CaCertificateUrl
			apiClientFactory := business.New(business.Config{CaCertificateUrl: caCertUrl})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				reg.GrpcApiAddress,
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
					GrpcApiAddress: reg.GrpcApiAddress,
					VpnAddress:     reg.VpnApiAddress,
				},
				apiGrpcClient,
				installer,
				sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress},
			).Run(ctx, startVpn.RunConfig{
				ProjectNameOrId:  args[0],
				Mtu:              params.GetUint32("mtu"),
				CaCertificateUrl: caCertUrl,
			})
		},
	}

	params.RegisterUInt32(cmd, "mtu", 1420, "vpn interface MTU")

	return cmd
}
