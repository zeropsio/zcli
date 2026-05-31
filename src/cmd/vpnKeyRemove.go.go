package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
)

func vpnKeyRemoveCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("remove").
		Short("Removes registered VPN public keys from project via API.").
		ScopeLevel(cmdBuilder.ScopeProject()).
		Arg("public-keys", cmdBuilder.OptionalArg(), cmdBuilder.ArrayArg(), cmdBuilder.OptionalArgLabel("[ public-keys ... ]")).
		HelpFlag("Help for the 'vpn key remove' command.").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			for _, publicKey := range cmdData.Args["public-keys"] {
				exists, err := repository.ProjectVpnKeyExists(ctx, cmdData.RestApiClient, project.Id, publicKey)
				if err != nil {
					return err
				}
				if !exists {
					cmdData.UxBlocks.PrintWarningTextf("Public key not found: %s", publicKey)
					continue
				}

				if err := repository.DeleteProjectVpnKey(ctx, cmdData.RestApiClient, project.Id, publicKey); err != nil {
					return err
				}

				cmdData.UxBlocks.PrintSuccessTextf("Removed registered public key from project via API: %s", publicKey)
			}

			return nil
		})
}
