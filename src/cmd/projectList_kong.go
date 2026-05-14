package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

// ProjectListCmd renders a table of all projects the user can see.
type ProjectListCmd struct{}

func (*ProjectListCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	return uxHelpers.PrintProjectList(ctx, auth.RestApiClient, d.Stdout)
}
