package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	paramsPackage "github.com/zerops-io/zcli/src/utils/params"
)

var (
	params *paramsPackage.Handler
)

var BuiltinToken string

func ExecuteCmd() error {
	params = paramsPackage.New()

	rootCmd := &cobra.Command{
		Use: "zcli",
	}

	params.RegisterString(rootCmd, "restApiAddress", "https://app.zerops.dev", "address of rest api", paramsPackage.Persistent())
	params.RegisterString(rootCmd, "grpcApiAddress", "app.zerops.dev:20902", "address of grpc api", paramsPackage.Persistent())
	params.RegisterString(rootCmd, "vpnApiAddress", "vpn.app.zerops.dev", "address of vpn api", paramsPackage.Persistent())

	rootCmd.AddCommand(deployCmd())
	rootCmd.AddCommand(vpnCmd())
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(logCmd())
	rootCmd.AddCommand(daemonCmd())

	err := params.InitViper()
	if err != nil {
		return err
	}

	err = rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func regSignals(contextCancel func()) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("\n", "signal:", sig)
		contextCancel()
	}()
}
