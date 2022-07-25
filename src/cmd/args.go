package cmd

import (
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func MinimumNArgs(num int) cobra.PositionalArgs {
	return customPositionalArgs(cobra.MinimumNArgs(num))
}

func ExactNArgs(num int) cobra.PositionalArgs {
	return customPositionalArgs(cobra.ExactArgs(num))
}

func customPositionalArgs(fn cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		err := fn(cmd, args)
		if err != nil {
			return fmt.Errorf("%w\nUsage: %s", err, buildUsage(cmd))
		}
		return nil
	}
}

func buildUsage(cmd *cobra.Command) string {
	parent := cmd.Parent()
	var parentUsage string
	if parent != nil {
		parentUsage = buildUsage(parent)
		parentUsage += " "
	}
	return parentUsage + cmd.Use
}

func helpText(add string) string {
	return i18n.DisplayHelp + add
}
