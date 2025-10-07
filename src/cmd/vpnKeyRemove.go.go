package cmd

import (
	"context"
	"encoding/base64"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/types"
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
				getResponse, err := cmdData.RestApiClient.GetProjectVpn(
					ctx,
					path.ProjectIdBase64PublicKey{
						Id:              project.Id,
						Base64PublicKey: types.String(base64.URLEncoding.EncodeToString([]byte(publicKey))),
					},
					query.GetProjectVpn{},
				)
				if err != nil {
					return err
				}
				if getResponse.Err() != nil {
					cmdData.UxBlocks.PrintWarningTextf("Public key not found: %s", publicKey)
					continue
				}

				deleteResponse, err := cmdData.RestApiClient.DeleteProjectVpn(
					ctx,
					path.ProjectId{Id: project.Id},
					body.PostProjectVpn{PublicKey: types.String(publicKey)},
				)
				if err != nil {
					return err
				}
				if err := deleteResponse.Err(); err != nil {
					return err
				}

				cmdData.UxBlocks.PrintSuccessTextf("Removed registered public key from project via API: %s", publicKey)
			}

			return nil
		})
}
