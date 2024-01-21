package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
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
		Short(i18n.T(i18n.CmdProjectImport)).
		Long(i18n.T(i18n.CmdProjectImportLong)).
		Arg(projectImportArgName).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			// TODO - janhajek client via flag
			// TODO - janhajek interactive selector of clients

			// TODO - janhajek config
			yamlContent, err := yamlReader.ReadContent(uxBlocks, cmdData.Args[projectImportArgName][0], "./")
			if err != nil {
				return err
			}

			importProjectResponse, err := cmdData.RestApiClient.PostProjectImport(
				ctx,
				body.ProjectImport{
					// TODO - janhajek client id
					ClientId: uuid.ClientId(cmdData.Args[projectImportArgName][0]),
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
						Id:                  process.Id,
						RunningMessage:      service.Name.String() + ": " + process.ActionName.String(),
						ErrorMessageMessage: service.Name.String() + ": " + process.ActionName.String(),
						SuccessMessage:      service.Name.String() + ": " + process.ActionName.String(),
					})
				}
			}

			uxBlocks.PrintLine(i18n.T(i18n.ServiceCount, len(responseOutput.ServiceStacks)))
			uxBlocks.PrintLine(i18n.T(i18n.QueuedProcesses, len(processes)))
			uxBlocks.PrintLine(i18n.T(i18n.CoreServices))

			err = uxHelpers.ProcessCheckWithSpinner(ctx, cmdData.UxBlocks, cmdData.RestApiClient, processes)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfoLine(i18n.T(i18n.ProjectImported))

			return nil
		})
}
