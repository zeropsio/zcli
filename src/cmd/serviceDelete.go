package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/path"
)

func serviceDeleteCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("delete").
		Short(i18n.T(i18n.CmdDescServiceDelete)).
		ScopeLevel(cmdBuilder.ScopeService()).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		BoolFlag("confirm", false, i18n.T(i18n.ConfirmFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpServiceDelete)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			question := styles.NewStringBuilder()
			question.WriteString("Service ")
			question.WriteStyledString(
				styles.SelectStyle().
					Bold(true),
				service.Name.String(),
			)
			question.WriteString(" will be deleted?")
			question.WriteString("\n")
			question.WriteString("Are you sure?")

			if !cmdData.Params.GetBool("confirm") {
				confirmed, err := uxHelpers.YesNoPrompt(
					ctx,
					question.String(),
					prompt.WithDialogBoxStyle(
						styles.DialogBox().
							BorderForeground(styles.ErrorColor),
					),
				)
				if err != nil {
					return err
				}
				if !confirmed {
					cmdData.UxBlocks.PrintInfoText(i18n.T(i18n.DestructiveOperationConfirmationFailed))
					return nil
				}
			}

			deleteServiceResponse, err := cmdData.RestApiClient.DeleteServiceStack(
				ctx,
				path.ServiceStackId{
					Id: service.ID,
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := deleteServiceResponse.Output()
			if err != nil {
				return err
			}

			processId := responseOutput.Id

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.ServiceDeleting),
					ErrorMessageMessage: i18n.T(i18n.ServiceDeleteFailed),
					SuccessMessage:      i18n.T(i18n.ServiceDeleted),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
