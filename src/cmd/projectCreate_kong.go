package cmd

import (
	"context"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// ProjectCreateCmd creates a new project under an org. Doesn't use the scope
// hook — it defines new state. Interactive selectors fire only when the
// process is TTY-attached and required inputs are missing.
type ProjectCreateCmd struct {
	Name         string   `name:"name" help:"Project name"`
	OrgID        string   `name:"org-id" help:"Organization ID to create project for"`
	Tags         []string `name:"tags" help:"Project tags. Comma separated list or repeated flag."`
	Out          string   `name:"out" help:"Output format of command, using golang's text/template engine."`
	Mode         string   `name:"mode" default:"LIGHT" help:"Project mode (LIGHT|SERIOUS)."`
	EnvIsolation string   `name:"env-isolation" default:"service" help:"Env isolation rule [service, none]."`
	SshIsolation string   `name:"ssh-isolation" default:"vpn" help:"SSH isolation rule."`
	Location     string   `name:"location" help:"Project primary/fallback location (e.g. eu-central)."`
}

func (c *ProjectCreateCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}

	mode := strings.ToUpper(c.Mode)
	if !enum.ProjectModeEnum(mode).Is(enum.ProjectModeEnumAllPublic()...) {
		return errors.Errorf("Invalid --mode, expected one of %s, got %s",
			enum.ProjectModeEnumAllPublic(), mode)
	}

	var outTemplate *template.Template
	if c.Out != "" {
		outTemplate, err = template.New("out").Parse(c.Out)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	var org entity.Org
	switch {
	case c.OrgID != "":
		org, err = repository.GetOrgById(ctx, auth.RestApiClient, uuid.ClientId(c.OrgID))
		if err != nil {
			return err
		}
	case !terminal.IsTerminal():
		return errors.New("Must specify organization ID with --org-id")
	default:
		org, err = uxHelpers.PrintOrgSelector(ctx, auth.RestApiClient, uxHelpers.WithOrgPickOnlyOneItem(true))
		if err != nil {
			return err
		}
	}
	d.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected org", org.Name.String()))

	name := c.Name
	if name == "" {
		if !terminal.IsTerminal() {
			return errors.New("Must specify name with --name")
		}
		label := styles.NewStringBuilder()
		label.WriteString("Type ")
		label.WriteStyledString(styles.SelectStyle().Bold(true), "project")
		label.WriteString(" name")
		name, err = uxBlock.Run(
			input.NewRoot(
				ctx,
				input.WithLabel(label.String()),
				input.WithHelpPlaceholder(),
				input.WithPlaceholderStyle(styles.HelpStyle()),
				input.WithoutPrompt(),
			),
			input.GetValueFunc,
		)
		if err != nil {
			return err
		}
	}

	postProject := entity.PostProject{
		OrgId:        org.Id,
		Name:         types.NewString(name),
		Tags:         c.Tags,
		Mode:         enum.ProjectModeEnum(mode),
		SshIsolation: types.NewStringNull(c.SshIsolation),
		EnvIsolation: types.NewStringNull(c.EnvIsolation),
	}
	if c.Location != "" {
		postProject.Location = stringId.NewLocationIdNullFromString(c.Location)
	} else if terminal.IsTerminal() {
		location, err := uxHelpers.PrintLocationSelector(ctx, auth.RestApiClient)
		if err != nil {
			return err
		}
		postProject.Location = location.Id.LocationIdNull()
		d.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected location", location.Name.String()))
	}

	project, err := repository.PostProject(ctx, auth.RestApiClient, postProject)
	if err != nil {
		return err
	}
	d.UxBlocks.PrintSuccessText("Project created")

	if outTemplate != nil {
		if err := outTemplate.Execute(d.Stdout, project); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
