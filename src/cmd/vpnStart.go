package cmd

import (
	"context"
	"errors"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/cliAction/startVpn"
	"github.com/zeropsio/zcli/src/daemonInstaller"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
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

			preferredPortMin, preferredPortMax, err := getMinMaxPort(params.GetString(cmd, "preferredPort"))
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
				PreferredPortMin: preferredPortMin,
				PreferredPortMax: preferredPortMax,
				CaCertificateUrl: caCertUrl,
			})
		},
	}

	params.RegisterUInt32(cmd, "mtu", 1420, i18n.MtuFlag)
	params.RegisterString(cmd, "preferredPort", "-", i18n.PreferredPortFlag)
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.VpnStartHelp))

	return cmd
}

func getMinMaxPort(in string) (uint32, uint32, error) {
	if in == "" {
		return 0, 0, nil
	}
	if in == "-" {
		return 0, 0, nil
	}
	parts := regexp.MustCompile("^([1-9][0-9]{0,4})?-([1-9][0-9]{0,4})?$").FindStringSubmatch(in)
	if len(parts) != 3 {
		return 0, 0, errors.New("invalid port range")
	}

	var min int
	var err error
	if parts[1] != "" {
		min, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
	}
	var max int
	if parts[2] != "" {
		max, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, 0, err
		}
	}
	return uint32(min), uint32(max), nil
}
