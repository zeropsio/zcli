package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

var support string

func supportCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("support").
		Short(i18n.T(i18n.CmdDescSupport)).
		HelpFlag(i18n.T(i18n.CmdHelpSupport)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {


            fmt.Println("You can contact Zerops support via:")
            fmt.Println("- E-mail:  team@zerops.io")
            fmt.Println("- Discord: https://discord.com/invite/WDvCZ54")
            fmt.Println(`
Additionally, you can explore our documentation
at https://docs.zerops.io/references/cli for further details.
            `)
            return nil

			return nil
		})
}

