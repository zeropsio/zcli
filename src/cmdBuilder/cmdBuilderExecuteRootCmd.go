package cmdBuilder

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/params"
)

type QuietMode int

const (
	QuietModeConfirmAll QuietMode = iota
	QuietModeConfirmOnlyDestructive
	QuietModeConfirmNothing
)

var QuietModeFlag int

type TerminalMode string

const (
	TerminalModeAuto     TerminalMode = "auto"
	TerminalModeDisabled              = "disabled"
	TerminalModeEnabled               = "enabled"
)

var TerminalFlag string

// Chicken-and-egg problem.
// I would like to log errors at one place after the execution of the root command.
// To do that, I need to know the log file path before the execution.
// To know the log file path, I need to parse the persistent flags.
// But these flags are parsed during the execution of the root command.
// So, I moved the logging inside the root command.
// This way, it logs everything. Except the unknown command error.
// This error needs to be handled here. Simple fmt.Println(err.Error()) is enough.
// But with this line, other errors are logged twice. Once here, once in the root command.
// So, I added a special error to skip the logging after the root command.
var skipErr = errors.New("skip")

func (b *CmdBuilder) CreateAndExecuteRootCobraCmd() error {
	rootCmd := createRootCommand()

	params := params.New()

	for _, cmd := range b.commands {
		cobraCmd, err := b.buildCobraCmd(cmd, params)
		if err != nil {
			return err
		}
		rootCmd.AddCommand(cobraCmd)
	}

	err := rootCmd.Execute()
	if err != nil {
		if !errors.Is(err, skipErr) {
			fmt.Println(err.Error())
		}
	}

	return nil
}

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "zcli",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		SilenceErrors:     true,
	}

	// FIXME - janhajek add a dynamic help for subcommands
	rootCmd.Flags().BoolP("help", "h", false, i18n.T(i18n.DisplayHelp)+i18n.T(i18n.GroupHelp))

	rootCmd.PersistentFlags().IntVar(&QuietModeFlag, "quiet", int(QuietModeConfirmAll), i18n.T(i18n.QuietModeFlag))
	rootCmd.PersistentFlags().StringVar(&TerminalFlag, "terminal", "auto", i18n.T(i18n.TerminalFlag))

	return rootCmd
}
