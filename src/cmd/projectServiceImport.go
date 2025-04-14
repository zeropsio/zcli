package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/types"
)

const serviceImportArgName = "importYamlPath"

func projectServiceImportCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("service-import").
		Short(i18n.T(i18n.CmdDescProjectServiceImport)).
		Long(i18n.T(i18n.CmdDescProjectServiceImportLong)).
		ScopeLevel(scope.Project).
		Arg(serviceImportArgName).
		HelpFlag(i18n.T(i18n.CmdHelpProjectServiceImport)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			if len(cmdData.Args[serviceImportArgName]) == 0 {
				return errors.New(i18n.T(i18n.ServiceImportYamlPathMissing))
			}

			yamlContent, err := yamlReader.ReadContent(uxBlocks, cmdData.Args[serviceImportArgName][0], "./")
			if err != nil {
				return errors.Wrap(err, i18n.T(i18n.ServiceImportYamlReadFailed))
			}

			importServiceResponse, err := cmdData.RestApiClient.PostServiceStackImport(
				ctx,
				body.ServiceStackImport{
					ProjectId: cmdData.Project.ID,
					Yaml:      types.Text(yamlContent),
				},
			)
			if err != nil {
				return errors.Wrap(err, i18n.T(i18n.ServiceImportFailed))
			}

			responseOutput, err := importServiceResponse.Output()
			if err != nil {
				return errors.Wrap(err, i18n.T(i18n.ServiceImportResponseParseFailed))
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
				return errors.Wrap(err, i18n.T(i18n.ServiceImportProcessCheckFailed))
			}

			uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.ServiceImported)))

			return nil
		})
}
