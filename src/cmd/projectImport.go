package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
)

const projectImportArgName = "import-yaml-path"

func projectImportCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("project-import").
		Short(i18n.T(i18n.CmdDescProjectImport)).
		Long(i18n.T(i18n.CmdDescProjectImportLong)).
		Arg(projectImportArgName).
		StringFlag("org-id", "", i18n.T(i18n.OrgIdFlag)).
		StringFlag("working-dir", "./", i18n.T(i18n.BuildWorkingDir)).
		HelpFlag(i18n.T(i18n.CmdHelpProjectImport)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			orgId, err := getOrgId(ctx, cmdData)
			if err != nil {
				return err
			}

			var yamlContent []byte
			if cmdData.Args[projectImportArgName][0] == "-" {
				yamlContent, err = yamlReader.ReadImportYamlContentFromStdin(uxBlocks)
				if err != nil {
					return err
				}
			} else {
				yamlContent, err = yamlReader.ReadImportYamlContent(
					uxBlocks,
					cmdData.Args[projectImportArgName][0],
					cmdData.Params.GetString("working-dir"),
				)
				if err != nil {
					return err
				}
			}

			importedServices, err := repository.ProjectImport(ctx, cmdData.RestApiClient, orgId, string(yamlContent))
			if err != nil {
				return err
			}

			var processes []uxHelpers.Process
			for _, service := range importedServices {
				for _, process := range service.Processes {
					processes = append(processes, uxHelpers.Process{
						F:                   uxHelpers.CheckZeropsProcess(process.Id, cmdData.RestApiClient, cmdData.UxBlocks),
						RunningMessage:      service.Name + ": " + process.ActionName,
						ErrorMessageMessage: service.Name + ": " + process.ActionName,
						SuccessMessage:      service.Name + ": " + process.ActionName,
					})
				}
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ServiceCount, len(importedServices))))
			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.QueuedProcesses, len(processes))))
			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.CoreServices)))

			err = uxHelpers.ProcessCheckWithSpinner(ctx, cmdData.UxBlocks, processes)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ProjectImported)))

			return nil
		})
}

func getOrgId(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) (entity.OrgId, error) {
	orgId := entity.OrgId(cmdData.Params.GetString("org-id"))
	if orgId != "" {
		return orgId, nil
	}

	orgs, err := repository.GetAllOrgs(ctx, cmdData.RestApiClient)
	if err != nil {
		return "", err
	}

	if len(orgs) == 1 {
		return orgs[0].Id, nil
	}

	selectedOrg, err := uxHelpers.PrintOrgSelector(ctx, cmdData.RestApiClient)
	if err != nil {
		return "", err
	}

	return selectedOrg.Id, nil
}
