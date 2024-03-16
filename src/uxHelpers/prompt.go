package uxHelpers

import (
	"context"

	"github.com/zeropsio/zcli/src/uxBlock"
)

func YesNoPrompt(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	questionMessage string,
) (bool, error) {
	// TODO - janhajek translate
	choices := []string{"NO", "YES"}
	choice, err := uxBlocks.Prompt(ctx, questionMessage, choices)
	if err != nil {
		return false, err
	}

	return choice == 1, nil
}
