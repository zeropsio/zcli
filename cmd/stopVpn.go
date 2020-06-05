package cmd

import (
	"context"

	"github.com/zerops-io/zcli/src/command/stopVpn"

	"github.com/spf13/cobra"
)

func stopVpnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "stopVpn",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel, logger)

			sudoers := createSudoers()

			storage, err := createStorage()
			if err != nil {
				return err
			}

			return stopVpn.New(
				stopVpn.Config{},
				logger,
				sudoers,
				storage,
			).Run(ctx, stopVpn.RunConfig{})
		},
	}

	return cmd
}
