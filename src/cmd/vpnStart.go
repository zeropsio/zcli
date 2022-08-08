package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/cliAction/startVpn"
	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
)

func vpnStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start projectNameOrId",
		Short:        i18n.CmdVpnStart,
		Long:         i18n.VpnStartLong,
		SilenceUsage: true,
		Args:         ExactNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}
			token, err := getToken(storage)
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

			caCertUrl := reg.CaCertificateUrl
			apiClientFactory := zBusinessZeropsApiProtocol.New(zBusinessZeropsApiProtocol.Config{CaCertificateUrl: caCertUrl})
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
				Token:            token,
				Mtu:              params.GetUint32("mtu"),
				CaCertificateUrl: caCertUrl,
			})
		},
	}

	params.RegisterUInt32(cmd, "mtu", 1420, i18n.MtuFlag)
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.VpnStartHelp))

	return cmd
}
