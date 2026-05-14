package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
)

// ProjectEnvCmd renders the scoped project's env file through a Go template,
// one line per var. Project resolved via the scope hook.
type ProjectEnvCmd struct {
	kongCmd.ProjectScope

	Export   bool   `name:"export" help:"Prepends export keyword to each env in output: 'export {{.Key}}={{.Value}}'."`
	Template string `name:"template" default:"{{.Key}}={{.Value}}" help:"Output template."`
	Service  string `name:"service" help:"Service name, in which context the environment variables are output."`
	UserOnly bool   `name:"user-only" help:"Exclude all non-user env variables."`
}

func (c *ProjectEnvCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	templateString := c.Template
	if c.Export {
		templateString = "export {{.Key}}={{.Value}}"
	}

	var serviceName types.EmptyString
	overrideEnvIsolation := enum.GetProjectEnvFileOverrideEnvIsolationEnumNone
	if c.Service != "" {
		serviceName = types.NewEmptyString(c.Service)
		overrideEnvIsolation = enum.GetProjectEnvFileOverrideEnvIsolationEnumService
	}

	tmpl, err := template.New("envs").Parse(templateString)
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := auth.RestApiClient.GetProjectEnvFile(
		ctx,
		path.ProjectId{Id: project.Id},
		query.GetProjectEnvFile{
			Name:                 serviceName,
			OverrideEnvIsolation: overrideEnvIsolation,
			UserOnly:             types.NewBool(c.UserOnly),
		},
	)
	if err != nil {
		return errors.WithStack(err)
	}
	envs, err := resp.Output()
	if err != nil {
		return errors.WithStack(err)
	}

	out := new(strings.Builder)
	scanner := bufio.NewScanner(strings.NewReader(envs.EnvFile.String()))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) != 2 {
			d.UxBlocks.PrintWarningText(fmt.Sprintf("unexpected env format: %s", scanner.Text()))
			continue
		}
		if err := tmpl.Execute(out, Env{Key: parts[0], Value: parts[1]}); err != nil {
			return errors.WithStack(err)
		}
		out.WriteRune('\n')
	}
	if out.Len() > 0 {
		d.Stdout.Println(out.String()[:out.Len()-1])
	}
	return nil
}
