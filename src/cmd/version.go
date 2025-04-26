package cmd

import (
	"context"
	"runtime"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	getVersion "github.com/zeropsio/zcli/src/version"
)

func versionCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("version").
		Short(i18n.T(i18n.CmdDescVersion)).
		HelpFlag(i18n.T(i18n.CmdHelpVersion)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			cmdData.Stdout.Printf("zcli version %s (%s) %s/%s\n", getVersion.GetCurrent(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
			getVersion.PrintVersionCheck(ctx, cmdData.Stdout)
			return nil
		})
}
