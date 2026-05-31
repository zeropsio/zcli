package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
)

func vpnKeyListCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("list").
		Short("Lists all registered VPN public keys of the project.").
		ScopeLevel(cmdBuilder.ScopeProject()).
		HelpFlag("Help for the 'vpn key list' command.").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			publicKeys, err := repository.ListProjectVpnPublicKeys(ctx, cmdData.RestApiClient, project.Id)
			if err != nil {
				return err
			}

			for _, publicKey := range publicKeys {
				cmdData.Stdout.Println(publicKey)
			}

			return nil
		})
}
