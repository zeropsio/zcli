package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

var version string

func versionCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("version").
		Short(i18n.T(i18n.CmdDescVersion)).
		HelpFlag(i18n.T(i18n.CmdHelpVersion)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			fmt.Printf("zcli version %s (%s) %s/%s\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)

			return nil
		})
}
