package cmd

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/i18n"
)

func daemonRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        i18n.CmdDaemonRun,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			logger, err := createLogger()
			if err != nil {
				return err
			}

			storage, err := createDaemonStorage()
			if err != nil {
				return err
			}

			wg := sync.WaitGroup{}

			dnsServer := createDnsServer()
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := dnsServer.Run(ctx)
				if err != nil {
					logger.Error(err)
					cancel()
				}
			}()

			vpn := createVpn(storage, dnsServer, logger)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := vpn.Run(ctx)
				if err != nil {
					logger.Error(err)
					cancel()
				}
			}()

			grpcServer := createDaemonGrpcServer(vpn)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := grpcServer.Run(ctx)
				if err != nil {
					logger.Error(err)
					cancel()
				}
			}()

			logger.Info("daemon is running")

			wg.Wait()

			logger.Info("daemon ended")

			return nil
		},
	}

	return cmd
}
