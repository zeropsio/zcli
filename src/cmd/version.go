package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/i18n"
)

var Version string

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "version",
		Short:        i18n.CmdVersion,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("zcli version %s (%s) %s/%s\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		},
	}

	return cmd
}
