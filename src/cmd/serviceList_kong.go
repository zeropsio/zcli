package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

// ServiceListCmd renders all services in the scoped project.
type ServiceListCmd struct {
	kongCmd.ProjectScope
}

func (c *ServiceListCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}
	return uxHelpers.PrintServiceList(ctx, auth.RestApiClient, d.Stdout, project)
}
