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
		RunE:         run,
	}

	cmd.PersistentFlags().BoolP("help", "h", false, helpText(i18n.DaemonRunHelp))
	return cmd
}

func daemonRun(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	regSignals(cancel)

	if err := prepareEnvironment(); err != nil {
		return err
	}

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

	vpnHandler := createVpn(storage, dnsServer, logger)

	grpcServer, err := createDaemonGrpcServer(vpnHandler)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Run(ctx)
		if err != nil {
			logger.Error(err)
			cancel()
		}
	}()

	vpnProlong := createVpnProlong(storage, logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := vpnProlong.Run(ctx)
		if err != nil {
			logger.Error(err)
			cancel()
		}
	}()

	logger.Info("daemon is running")

	wg.Wait()

	logger.Info("daemon ended")

	return nil
}
