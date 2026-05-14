package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
)

// ProjectServiceImportCmd imports one or more services into an existing
// project from a yaml spec. Positional `<import-yaml-path>` can be `-` for
// stdin. Project is resolved via the scope hook.
type ProjectServiceImportCmd struct {
	kongCmd.ProjectScope

	ImportYamlPath string `arg:"" name:"import-yaml-path" help:"Path to the import YAML, or '-' for stdin."`
}

func (c *ProjectServiceImportCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}

	var yamlContent []byte
	if c.ImportYamlPath == "-" {
		yamlContent, err = yamlReader.ReadImportYamlContentFromStdin(d.UxBlocks)
	} else {
		yamlContent, err = yamlReader.ReadImportYamlContent(d.UxBlocks, c.ImportYamlPath, "./")
	}
	if err != nil {
		return err
	}

	resp, err := auth.RestApiClient.PostProjectServiceStackImport(
		ctx,
		path.ProjectId{Id: project.Id},
		body.ServiceStackImport{Yaml: types.Text(yamlContent)},
	)
	if err != nil {
		return err
	}
	out, err := resp.Output()
	if err != nil {
		return err
	}

	var processes []uxHelpers.Process
	for _, service := range out.ServiceStacks {
		for _, p := range service.Processes {
			processes = append(processes, uxHelpers.Process{
				F:                   uxHelpers.CheckZeropsProcess(p.Id, auth.RestApiClient, d.UxBlocks),
				RunningMessage:      service.Name.String() + ": " + p.ActionName.String(),
				ErrorMessageMessage: service.Name.String() + ": " + p.ActionName.String(),
				SuccessMessage:      service.Name.String() + ": " + p.ActionName.String(),
			})
		}
	}

	d.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ServiceCount, len(out.ServiceStacks))))
	d.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.QueuedProcesses, len(processes))))

	if err := uxHelpers.ProcessCheckWithSpinner(ctx, d.UxBlocks, processes); err != nil {
		return err
	}
	d.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ServiceImported)))
	return nil
}
