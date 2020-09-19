package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/daemonInstaller"
	"github.com/zerops-io/zcli/src/i18n"
)

func daemonInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        i18n.CmdDaemonInstall,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			installer, err := daemonInstaller.New(daemonInstaller.Config{})
			if err != nil {
				return err
			}

			err = installer.Install()
			if err != nil {
				return err
			}

			fmt.Println(i18n.DaemonInstallSuccess)

			return nil
		},
	}

	return cmd
}
