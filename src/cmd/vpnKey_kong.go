package cmd

import (
	"context"
	"encoding/base64"

	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/types"
)

// VpnKeyCmd groups vpn key subcommands.
type VpnKeyCmd struct {
	List   VpnKeyListCmd   `cmd:"" help:"List all registered VPN public keys of the project."`
	Remove VpnKeyRemoveCmd `cmd:"" help:"Remove registered VPN public keys from the project."`
}

// VpnKeyListCmd prints the project's registered public keys, one per line.
type VpnKeyListCmd struct {
	kongCmd.ProjectScope
}

func (c *VpnKeyListCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}
	listResp, err := auth.RestApiClient.GetProjectVpnList(ctx, path.ProjectId{Id: project.Id}, query.GetProjectVpn{})
	if err != nil {
		return err
	}
	list, err := listResp.Output()
	if err != nil {
		return err
	}
	for _, peer := range list.Peers {
		d.Stdout.Println(peer.PublicKey)
	}
	return nil
}

// VpnKeyRemoveCmd revokes the given public keys from the project.
type VpnKeyRemoveCmd struct {
	kongCmd.ProjectScope

	PublicKeys []string `arg:"" name:"public-keys" optional:"" help:"Public keys to remove."`
}

func (c *VpnKeyRemoveCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	for _, publicKey := range c.PublicKeys {
		getResp, err := auth.RestApiClient.GetProjectVpn(
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
		if getResp.Err() != nil {
			d.UxBlocks.PrintWarningTextf("Public key not found: %s", publicKey)
			continue
		}
		deleteResp, err := auth.RestApiClient.DeleteProjectVpn(
			ctx,
			path.ProjectId{Id: project.Id},
			body.PostProjectVpn{PublicKey: types.String(publicKey)},
		)
		if err != nil {
			return err
		}
		if err := deleteResp.Err(); err != nil {
			return err
		}
		d.UxBlocks.PrintSuccessTextf("Removed registered public key from project via API: %s", publicKey)
	}
	return nil
}
