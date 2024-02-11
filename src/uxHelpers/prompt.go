package uxHelpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
)

func YesNoPromptDestructive(ctx context.Context, uxBlocks uxBlock.UxBlocks, message string) error {
	// TODO - janhajek translate
	choices := []string{"NO", "YES"}
	choice, err := uxBlocks.Prompt(ctx, message, choices)
	if err != nil {
		return err
	}

	if choice == 0 {
		return errors.New(i18n.T(i18n.DestructiveOperationConfirmationFailed))
	}

	return nil
}
