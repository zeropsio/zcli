package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/output"
)

func sudoCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("sudo").
		Short(i18n.T(i18n.CmdDescSudo)).
		HelpFlag(i18n.T(i18n.CmdHelpSudo)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			mode, err := getSudoMode(ctx, cmdData)
			if err != nil {
				return err
			}
			if mode.InSudo.Native() {
				printSudoOn(cmdData, mode)
				return nil
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.SudoOff)))
			if !cmdData.IsLoginSession() {
				return errorsx.NewUserError(i18n.T(i18n.SudoTokenWithout), nil)
			}
			if terminal.IsInteractive() {
				confirmed, err := uxHelpers.YesNoPrompt(ctx, i18n.T(i18n.SudoEnablePrompt), prompt.WithCursorPosition(1))
				if err != nil || !confirmed {
					return err
				}
			}
			if err := cmdData.EnableSudo(ctx); err != nil {
				return err
			}

			mode, err = getSudoMode(ctx, cmdData)
			if err != nil {
				return err
			}
			if !mode.InSudo.Native() {
				return errorsx.NewUserError(i18n.T(i18n.SudoNotApplied), nil)
			}
			printSudoOn(cmdData, mode)
			return nil
		})
}

func getSudoMode(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) (output.SudoMode, error) {
	response, err := cmdData.RestApiClient.GetAuthSudo(ctx)
	if err != nil {
		return output.SudoMode{}, err
	}
	return response.Output()
}

// printSudoOn shows the expiry only for a login session, a token with sudo mode keeps it as long as the token is valid.
func printSudoOn(cmdData *cmdBuilder.LoggedUserCmdData, mode output.SudoMode) {
	until, ok := mode.SudoUntil.Get()
	if !cmdData.IsLoginSession() || !ok {
		cmdData.UxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.SudoOn)))
		return
	}
	cmdData.UxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.SudoOnFor, remaining(time.Until(until.Native())))))
}

func remaining(d time.Duration) string {
	minutes := max(int(d.Round(time.Minute)/time.Minute), 1)
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh %dm", minutes/60, minutes%60)
}
