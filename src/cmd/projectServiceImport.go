package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
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

			var jobs []uxBlock.SpinnerJob
			for _, service := range responseOutput.ServiceStacks {
				for _, process := range service.Processes {
					label := service.Name.String() + ": " + process.ActionName.String()
					pid := process.Id
					jobs = append(jobs, uxBlock.SpinnerJob{
						Spec: uxBlock.SpinnerSpec{
							Message:        label,
							SuccessMessage: label,
							ErrorMessage:   label,
						},
						Run: func(_ uxBlock.SpinnerHandle) error {
							return uxHelpers.PollProcess(ctx, cmdData.RestApiClient, cmdData.UxBlocks, pid, nil)
						},
					})
				}
			}

			uxBlocks.Print(styles.InfoLine, i18n.T(i18n.ServiceCount, len(responseOutput.ServiceStacks)))
			uxBlocks.Print(styles.InfoLine, i18n.T(i18n.QueuedProcesses, len(jobs)))

			if err := cmdData.UxBlocks.WithSpinners(ctx, jobs); err != nil {
				return err
			}

			uxBlocks.Print(styles.InfoLine, i18n.T(i18n.ServiceImported))

			return nil
		})
}
