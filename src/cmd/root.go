package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zerops-io/zcli/src/constants"

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

	params.RegisterPersistentString(rootCmd, constants.PersistentParamRestApiAddress, "https://api.app.zerops.io", "address of rest api")
	params.RegisterPersistentString(rootCmd, constants.PersistentParamGrpcApiAddress, "api.app.zerops.io:20902", "address of grpc api")
	params.RegisterPersistentString(rootCmd, constants.PersistentParamVpnApiAddress, "vpn.app.zerops.io", "address of vpn api")
	params.RegisterPersistentString(rootCmd, constants.PersistentParamCaCertificateUrl, "https://api.app.zerops.io/ca.crt", "download url for certificate of Zerops certificate authority used for tls encrypted communication via gRPC")

	rootCmd.AddCommand(deployCmd())
	rootCmd.AddCommand(pushCmd())
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
