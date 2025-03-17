package cmdBuilder

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/uxBlock"
)

func buildCobraCmd(
	cmd *Cmd,
	flagParams *flagParams.Handler,
	uxBlocks *uxBlock.Blocks,
	cliStorage *cliStorage.Handler,
) (*cobra.Command, error) {
	cobraCmd := &cobra.Command{
		Short:         cmd.short,
		SilenceUsage:  cmd.silenceUsage,
		SilenceErrors: cmd.silenceError,
	}

	if cmd.helpTemplate != "" {
		cobraCmd.SetHelpTemplate(cmd.helpTemplate)
	}

	argNames := make([]string, len(cmd.args))
	for i, arg := range cmd.args {
		argName := arg.name
		if arg.optionalLabel != "" {
			argName = arg.optionalLabel
		}
		if arg.optional {
			argName = "[" + argName + "]"
		}
		argNames[i] = argName
	}
	cobraCmd.Use = strings.Join(append([]string{cmd.use}, argNames...), " ")

	if cmd.scopeLevel != nil {
		cmd.scopeLevel.AddCommandFlags(cmd)
	}

	for _, flag := range cmd.flags {
		flagSet := cobraCmd.Flags()
		switch defaultValue := flag.defaultValue.(type) {
		case string:
			flagSet.StringP(flag.name, flag.shorthand, defaultValue, flag.description)
		case int:
			flagSet.IntP(flag.name, flag.shorthand, defaultValue, flag.description)
		case bool:
			flagSet.BoolP(flag.name, flag.shorthand, defaultValue, flag.description)
		default:
			panic(fmt.Sprintf("unexpected type %T", flag.defaultValue))
		}

		if flag.hidden {
			err := flagSet.MarkHidden(flag.name)
			if err != nil {
				return nil, err
			}
		}
	}

	if cmd.guestRunFunc != nil || cmd.loggedUserRunFunc != nil {
		cobraCmd.RunE = createCmdRunFunc(cmd, flagParams, uxBlocks, cliStorage)
	}

	for _, childrenCmd := range cmd.childrenCmds {
		cobraChildrenCmd, err := buildCobraCmd(childrenCmd, flagParams, uxBlocks, cliStorage)
		if err != nil {
			return nil, err
		}
		cobraCmd.AddCommand(cobraChildrenCmd)
	}

	return cobraCmd, nil
}
