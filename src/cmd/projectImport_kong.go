package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// ProjectImportCmd creates a project (and its services) from a yaml spec.
// Positional `<import-yaml-path>` can be `-` for stdin. Doesn't use the
// scope hook — it's defining new state.
type ProjectImportCmd struct {
	ImportYamlPath string `arg:"" name:"import-yaml-path" help:"Path to the import YAML, or '-' for stdin."`
	OrgID          string `name:"org-id" help:"Org ID to import under."`
	WorkingDir     string `name:"working-dir" default:"./" help:"Working directory for relative paths."`
}

func (c *ProjectImportCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}

	orgID, err := resolveOrgID(ctx, auth.RestApiClient, c.OrgID)
	if err != nil {
		return err
	}

	var yamlContent []byte
	if c.ImportYamlPath == "-" {
		yamlContent, err = yamlReader.ReadImportYamlContentFromStdin(d.UxBlocks)
	} else {
		yamlContent, err = yamlReader.ReadImportYamlContent(d.UxBlocks, c.ImportYamlPath, c.WorkingDir)
	}
	if err != nil {
		return err
	}

	resp, err := auth.RestApiClient.PostClientProjectImport(
		ctx,
		path.ClientId{Id: orgID},
		body.ProjectImport{Yaml: types.Text(yamlContent)},
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
	d.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.CoreServices)))

	if err := uxHelpers.ProcessCheckWithSpinner(ctx, d.UxBlocks, processes); err != nil {
		return err
	}
	d.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ProjectImported)))
	return nil
}

// resolveOrgID picks an org by id (flag), by uniqueness (one org → no choice),
// or interactively. Shared with project-service-import.
func resolveOrgID(ctx context.Context, client *zeropsRestApiClient.Handler, orgIDFlag string) (uuid.ClientId, error) {
	if orgIDFlag != "" {
		return uuid.ClientId(orgIDFlag), nil
	}
	orgs, err := repository.GetAllOrgs(ctx, client)
	if err != nil {
		return "", err
	}
	if len(orgs) == 1 {
		return orgs[0].Id, nil
	}
	selected, err := uxHelpers.PrintOrgSelector(ctx, client)
	if err != nil {
		return "", err
	}
	return selected.Id, nil
}
