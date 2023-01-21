package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/support"
	paramsPackage "github.com/zeropsio/zcli/src/utils/params"
)

var (
	params *paramsPackage.Handler
)

var BuiltinToken string

func ExecuteCmd() error {
	params = paramsPackage.New()

	rootCmd := &cobra.Command{
		Use:               "zcli",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		SilenceErrors:     true,
	}

	rootCmd.AddCommand(initCommand())
	rootCmd.AddCommand(deployCmd())
	rootCmd.AddCommand(pushCmd())
	rootCmd.AddCommand(vpnCmd())
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(logCmd())
	rootCmd.AddCommand(daemonCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(regionList())
	rootCmd.AddCommand(projectCmd())
	rootCmd.AddCommand(serviceCmd())
	rootCmd.AddCommand(bucketCmd())

	rootCmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))

	err := params.InitViper()
	if err != nil {
		return err
	}

	ctx := support.Context(context.Background())
	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func regSignals(contextCancel func()) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		contextCancel()
	}()
}
