package cmdBuilder

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/params"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

type TerminalMode string

const (
	TerminalModeAuto     TerminalMode = "auto"
	TerminalModeDisabled TerminalMode = "disabled"
	TerminalModeEnabled  TerminalMode = "enabled"
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
var errSkipErrorReporting = errors.New("skipErrorReporting")

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
		if !errors.Is(err, errSkipErrorReporting) {
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

	rootCmd.PersistentFlags().StringVar(&TerminalFlag, "terminal", "auto", i18n.T(i18n.TerminalFlag))

	rootCmd.SetHelpTemplate(`` + styles.CobraSectionColor().SetString("Usage:").String() + `{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + styles.CobraSectionColor().SetString("Aliases:").String() + `
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

` + styles.CobraSectionColor().SetString("Examples:").String() + `
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

` + styles.CobraSectionColor().SetString("Available Commands:").String() + `{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

` + styles.CobraSectionColor().SetString("Additional Commands:").String() + `{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + styles.CobraSectionColor().SetString("Flags:").String() + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + styles.CobraSectionColor().SetString("Global Flags:").String() + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + styles.CobraSectionColor().SetString("Additional help topics:").String() + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .CommandPath .CommandPathPadding}}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	return rootCmd
}
