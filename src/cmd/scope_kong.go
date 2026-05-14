package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// ScopeCmd groups the scope subcommands. No Run on the parent — kong allows
// command-only nodes that just route to children.
type ScopeCmd struct {
	Project ScopeProjectCmd `cmd:"" help:"Sets the scope for project. All commands that require project ID will use the selected one."`
	Reset   ScopeResetCmd   `cmd:"" help:"Resets the scope for project and service."`
}

// ScopeProjectCmd sets the persisted project scope. Takes an optional
// project-id positional; without it, runs the interactive project selector.
type ScopeProjectCmd struct {
	ProjectID string `arg:"" name:"project-id" optional:"" help:"Project ID to scope to."`
}

func (c *ScopeProjectCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}

	if prevID, set := d.CliStorage.Data().ScopeProjectId.Get(); set {
		if prev, err := repository.GetProjectById(ctx, auth.RestApiClient, prevID); err != nil {
			if errorsx.Is(err, errorsx.ErrorCode(errorCode.ProjectNotFound)) {
				d.UxBlocks.PrintWarning(styles.WarningLine(err.Error()))
			} else {
				return err
			}
		} else {
			d.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.PreviouslyScopedProject), prev.Name.String()))
		}
	}

	var project entity.Project
	if c.ProjectID != "" {
		project, err = repository.GetProjectById(ctx, auth.RestApiClient, uuid.ProjectId(c.ProjectID))
		if err != nil {
			return err
		}
	} else {
		picked, err := uxHelpers.PrintProjectSelector(ctx, auth.RestApiClient)
		if err != nil {
			return err
		}
		project, err = picked.Expect("project is null")
		if err != nil {
			return err
		}
	}

	if _, err := d.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.ScopeProjectId = project.Id.ProjectIdNull()
		return data
	}); err != nil {
		return err
	}

	d.UxBlocks.PrintInfo(styles.InfoWithValueLine(i18n.T(i18n.SelectedProject), project.Name.String()))
	return nil
}

// ScopeResetCmd clears the persisted project scope.
type ScopeResetCmd struct{}

func (*ScopeResetCmd) Run(_ context.Context, d *kongCmd.CmdData) error {
	if _, err := d.RequireAuth(); err != nil {
		return err
	}
	if _, err := d.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.ScopeProjectId = uuid.ProjectIdNull{}
		return data
	}); err != nil {
		return errors.WithStack(err)
	}
	d.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ScopeReset)))
	return nil
}
