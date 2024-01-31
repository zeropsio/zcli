package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func YesNoPromptDestructive(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData, message string) error {
	// TODO - janhajek translate
	choices := []string{"NO", "YES"}
	choice, err := cmdData.UxBlocks.Prompt(ctx, message, choices)
	if err != nil {
		return err
	}

	if choice == 0 {
		return errors.New(i18n.T(i18n.DestructiveOperationConfirmationFailed))
	}

	return nil
}
