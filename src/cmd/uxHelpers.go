package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
)

func YesNoPromptDestructive(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData, message string) (bool, error) {
	if cmdData.QuietMode == cmdBuilder.QuietModeConfirmNothing {
		return true, nil
	}

	// TODO - janhajek translate
	choices := []string{"no", "yes"}
	choice, err := cmdData.UxBlocks.Prompt(ctx, message, choices)
	if err != nil {
		return false, err
	}

	return choice == 1, nil
}

func YesNoPromptNonDestructive(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData, message string) (bool, error) {
	if cmdData.QuietMode == cmdBuilder.QuietModeConfirmNothing || cmdData.QuietMode == cmdBuilder.QuietModeConfirmOnlyDestructive {
		return true, nil
	}

	// TODO - janhajek translate
	choices := []string{"no", "yes"}
	choice, err := cmdData.UxBlocks.Prompt(ctx, message, choices)
	if err != nil {
		return false, err
	}

	return choice == 1, nil
}
