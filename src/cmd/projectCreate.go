package cmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/sdk"
)

func projectCreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short(i18n.T(i18n.CmdDescProjectCreate)).
		ScopeLevel(scope.Project).
		Arg(scope.ProjectArgName, cmdBuilder.OptionalArg()).
		BoolFlag("confirm", false, i18n.T(i18n.ConfirmFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpProjectCreate)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			if !cmdData.Params.GetBool("confirm") {
				confirmed, err := uxHelpers.YesNoPrompt(
					ctx,
					cmdData.UxBlocks,
					i18n.T(i18n.ProjectCreateConfirm, cmdData.Project.Name),
				)
				if err != nil {
					return err
				}
				if !confirmed {
					return errors.New(i18n.T(i18n.DestructiveOperationConfirmationFailed))
				}
			}

			sdk.Handler{}.PostProject(ctx, body.PostProject{
				Name: cmdData.Project.Name,
			})

			return nil
		})
}
