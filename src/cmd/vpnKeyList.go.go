package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
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

			listResponse, err := cmdData.RestApiClient.GetProjectVpnList(
				ctx,
				path.ProjectId{Id: project.Id},
				query.GetProjectVpn{},
			)
			if err != nil {
				return err
			}
			list, err := listResponse.Output()
			if err != nil {
				return err
			}

			for _, peer := range list.Peers {
				cmdData.Stdout.Println(peer.PublicKey)
			}

			return nil
		})
}
