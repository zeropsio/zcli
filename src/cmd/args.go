package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func CustomMessageArgs(fn cobra.PositionalArgs, customMessage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		err := fn(cmd, args)
		if err != nil {
			return errors.New(customMessage)
		}
		return nil
	}
}
