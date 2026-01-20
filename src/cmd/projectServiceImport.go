package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
)

const serviceImportArgName = "import-yaml-path"

func projectServiceImportCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("service-import").
		Short(i18n.T(i18n.CmdDescProjectServiceImport)).
		ScopeLevel(cmdBuilder.ScopeProject()).
		Arg(serviceImportArgName).
		HelpFlag(i18n.T(i18n.CmdHelpProjectServiceImport)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks
			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}

			var yamlContent []byte
			if cmdData.Args[serviceImportArgName][0] == "-" {
				yamlContent, err = yamlReader.ReadImportYamlContentFromStdin(uxBlocks)
				if err != nil {
					return err
				}
			} else {
				yamlContent, err = yamlReader.ReadImportYamlContent(
					uxBlocks,
					cmdData.Args[serviceImportArgName][0],
					"./",
				)
				if err != nil {
					return err
				}
			}

			importServiceResponse, err := cmdData.RestApiClient.PostProjectServiceStackImport(
				ctx,
				path.ProjectId{Id: project.Id},
				body.ServiceStackImport{
					Yaml: types.Text(yamlContent),
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := importServiceResponse.Output()
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

			err = uxHelpers.ProcessCheckWithSpinner(ctx, cmdData.UxBlocks, processes)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ServiceImported)))

			return nil
		})
}
