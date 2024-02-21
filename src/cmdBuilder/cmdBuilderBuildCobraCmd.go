package cmdBuilder

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/params"
	"github.com/zeropsio/zcli/src/uxBlock"
)

func (b *CmdBuilder) buildCobraCmd(
	cmd *Cmd,
	params *params.Handler,
	uxBlocks uxBlock.UxBlocks,
	cliStorage *cliStorage.Handler,
) (*cobra.Command, error) {
	cobraCmd := &cobra.Command{
		Short:        cmd.short,
		SilenceUsage: cmd.silenceUsage,
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

	for _, dep := range getDependencyListFromRoot(cmd.scopeLevel) {
		dep.AddCommandFlags(cmd)
	}

	for _, flag := range cmd.flags {
		switch defaultValue := flag.defaultValue.(type) {
		case string:
			params.RegisterString(cobraCmd, flag.name, flag.shorthand, defaultValue, flag.description)
		case int:
			params.RegisterInt(cobraCmd, flag.name, flag.shorthand, defaultValue, flag.description)
		case bool:
			params.RegisterBool(cobraCmd, flag.name, flag.shorthand, defaultValue, flag.description)
		default:
			panic(fmt.Sprintf("unexpected type %T", flag.defaultValue))
		}

		if flag.hidden {
			err := cobraCmd.Flags().MarkHidden(flag.name)
			if err != nil {
				return nil, err
			}
		}
	}

	if cmd.guestRunFunc != nil || cmd.loggedUserRunFunc != nil {
		cobraCmd.RunE = b.createCmdRunFunc(cmd, params, uxBlocks, cliStorage)
	}

	for _, childrenCmd := range cmd.childrenCmds {
		cobraChildrenCmd, err := b.buildCobraCmd(childrenCmd, params, uxBlocks, cliStorage)
		if err != nil {
			return nil, err
		}
		cobraCmd.AddCommand(cobraChildrenCmd)
	}

	return cobraCmd, nil
}
