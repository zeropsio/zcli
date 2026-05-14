package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"
)

// ServiceStartCmd starts the scoped service.
type ServiceStartCmd struct {
	kongCmd.ServiceScope
}

func (c *ServiceStartCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	service, err := auth.RequireService(ctx, c.ServiceScope)
	if err != nil {
		return err
	}
	resp, err := auth.RestApiClient.PutServiceStackStart(ctx, path.ServiceStackId{Id: service.Id})
	if err != nil {
		return err
	}
	out, err := resp.Output()
	if err != nil {
		return err
	}
	return uxHelpers.ProcessCheckWithSpinner(ctx, d.UxBlocks, []uxHelpers.Process{{
		F:                   uxHelpers.CheckZeropsProcess(out.Id, auth.RestApiClient, d.UxBlocks),
		RunningMessage:      i18n.T(i18n.ServiceStarting),
		ErrorMessageMessage: i18n.T(i18n.ServiceStartFailed),
		SuccessMessage:      i18n.T(i18n.ServiceStarted),
	}})
}
