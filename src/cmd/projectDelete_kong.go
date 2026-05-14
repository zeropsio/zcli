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

// ProjectDeleteCmd deletes the project resolved via Auth.RequireProject.
// Optional positional `[project-id]` or `--project-id` flag; falls back to
// persisted scope or interactive selector when both are absent.
type ProjectDeleteCmd struct {
	kongCmd.ProjectScope

	ProjectArg string `arg:"" name:"project-id" optional:"" help:"Project ID to delete."`
	Confirm    bool   `name:"confirm" help:"Skip the destructive-operation confirmation prompt."`
}

func (c *ProjectDeleteCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	if c.ProjectScope.ProjectID == "" && c.ProjectArg != "" {
		c.ProjectScope.ProjectID = c.ProjectArg
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	if !c.Confirm {
		q := styles.NewStringBuilder()
		q.WriteString("Project ")
		q.WriteStyledString(styles.SelectStyle().Bold(true), project.Name.String())
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

	deleteResp, err := auth.RestApiClient.DeleteProject(ctx, path.ProjectId{Id: project.Id})
	if err != nil {
		return err
	}
	out, err := deleteResp.Output()
	if err != nil {
		return err
	}

	return uxHelpers.ProcessCheckWithSpinner(
		ctx,
		d.UxBlocks,
		[]uxHelpers.Process{{
			F:                   uxHelpers.CheckZeropsProcess(out.Id, auth.RestApiClient, d.UxBlocks),
			RunningMessage:      i18n.T(i18n.ProjectDeleting),
			ErrorMessageMessage: i18n.T(i18n.ProjectDeleteFailed),
			SuccessMessage:      i18n.T(i18n.ProjectDeleted),
		}},
	)
}
