package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

const projectImportArgName = "importYamlPath"

func projectImportCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("project-import").
		Short(i18n.T(i18n.CmdDescProjectImport)).
		Long(i18n.T(i18n.CmdDescProjectImportLong)).
		Arg(projectImportArgName).
		StringFlag("orgId", "", i18n.T(i18n.OrgIdFlag)).
		StringFlag("workingDir", "./", i18n.T(i18n.BuildWorkingDir)).
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
					cmdData.Params.GetString("workingDir"),
				)
				if err != nil {
					return err
				}
			}

			importProjectResponse, err := cmdData.RestApiClient.PostProjectImport(
				ctx,
				body.ProjectImport{
					ClientId: orgId,
					Yaml:     types.Text(yamlContent),
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := importProjectResponse.Output()
			if err != nil {
				return err
			}

			var processes []uxHelpers.Process
			for _, service := range responseOutput.ServiceStacks {
				for _, process := range service.Processes {
					processes = append(processes, uxHelpers.Process{
						F:                   uxHelpers.CheckZeropsProcess(process.Id, cmdData.RestApiClient),
						RunningMessage:      service.Name.String() + ": " + process.ActionName.String(),
						ErrorMessageMessage: service.Name.String() + ": " + process.ActionName.String(),
						SuccessMessage:      service.Name.String() + ": " + process.ActionName.String(),
					})
				}
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ServiceCount, len(responseOutput.ServiceStacks))))
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

func getOrgId(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) (uuid.ClientId, error) {
	orgId := uuid.ClientId(cmdData.Params.GetString("orgId"))
	if orgId != "" {
		return orgId, nil
	}

	orgs, err := repository.GetAllOrgs(ctx, cmdData.RestApiClient)
	if err != nil {
		return "", err
	}

	if len(orgs) == 1 {
		return orgs[0].ID, nil
	}

	selectedOrg, err := uxHelpers.PrintOrgSelector(ctx, cmdData.RestApiClient)
	if err != nil {
		return "", err
	}

	return selectedOrg.ID, nil
}
