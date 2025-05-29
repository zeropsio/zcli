package uxHelpers

import (
	"context"

	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
)

func YesNoPrompt(
	ctx context.Context,
	question string,
	opts ...prompt.Option,
) (bool, error) {
	choice, err := uxBlock.Run(
		prompt.NewRoot(
			ctx,
			question,
			[]string{"NO", "YES"},
			opts...,
		),
		prompt.GetChoiceCursor,
	)
	if err != nil {
		return false, err
	}

	return choice == 1, nil
}
