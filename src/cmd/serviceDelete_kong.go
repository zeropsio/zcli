package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"
)

// ServiceDeleteCmd deletes the scoped service.
type ServiceDeleteCmd struct {
	kongCmd.ServiceScope

	Confirm bool `name:"confirm" help:"Skip the destructive-operation confirmation prompt."`
}

func (c *ServiceDeleteCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	service, err := auth.RequireService(ctx, c.ServiceScope)
	if err != nil {
		return err
	}

	if !c.Confirm {
		q := styles.NewStringBuilder()
		q.WriteString("Service ")
		q.WriteStyledString(styles.SelectStyle().Bold(true), service.Name.String())
		q.WriteString(" will be deleted?\nAre you sure?")

		confirmed, err := uxHelpers.YesNoPrompt(ctx, q.String(),
			prompt.WithDialogBoxStyle(styles.DialogBox().BorderForeground(styles.ErrorColor)),
		)
		if err != nil {
			return err
		}
		if !confirmed {
			d.UxBlocks.PrintInfoText(i18n.T(i18n.DestructiveOperationConfirmationFailed))
			return nil
		}
	}

	resp, err := auth.RestApiClient.DeleteServiceStack(ctx, path.ServiceStackId{Id: service.Id})
	if err != nil {
		return err
	}
	out, err := resp.Output()
	if err != nil {
		return err
	}

	return uxHelpers.ProcessCheckWithSpinner(ctx, d.UxBlocks, []uxHelpers.Process{{
		F:                   uxHelpers.CheckZeropsProcess(out.Id, auth.RestApiClient, d.UxBlocks),
		RunningMessage:      i18n.T(i18n.ServiceDeleting),
		ErrorMessageMessage: i18n.T(i18n.ServiceDeleteFailed),
		SuccessMessage:      i18n.T(i18n.ServiceDeleted),
	}})
}
