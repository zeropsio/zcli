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

	rootCmd.AddCommand(importCmd())
	rootCmd.AddCommand(deployCmd())
	rootCmd.AddCommand(pushCmd())
	rootCmd.AddCommand(vpnCmd())
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(logCmd())
	rootCmd.AddCommand(daemonCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(deleteCmd())

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
