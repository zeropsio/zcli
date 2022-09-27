package cmd

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/i18n"
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

func daemonRun(ctx context.Context) error {
	cancelCtx, cancel := context.WithCancel(ctx)
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

	var wg sync.WaitGroup

	dnsServer := createDnsServer(logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := dnsServer.Run(cancelCtx)
		if err != nil {
			logger.Error(err)
			cancel()
		}
	}()

	vpnHandler := createVpn(storage, dnsServer, logger)

	if err := vpnHandler.ReloadVpn(cancelCtx); err != nil {
		return err
	}

	grpcServer, err := createDaemonGrpcServer(vpnHandler)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Run(cancelCtx)
		if err != nil {
			logger.Error(err)
			cancel()
		}
	}()

	vpnProlong := createVpnProlong(storage, logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := vpnProlong.Run(cancelCtx)
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
